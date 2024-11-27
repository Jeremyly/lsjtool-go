/*
 * @Author: lsjweiyi
 * @Date: 2024-05-02 17:42:58
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-02 22:00:04
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package service

import (
	"context"
	"encoding/json"
	"src/app/dto"
	"src/app/model"
	"src/app/service/outApi"
	"src/app/service/payService"
	g "src/global"
	"src/utils/base"
	"time"
)

var DocConvertTask = make(map[string]int, 0) // 记录任务进度

// taskListMap:用于记录任务是否结果，防止并发重复处理同一任务
func DocImgConvertRequest(task *model.ImgTask, taskListMap map[string]bool) {
	// 之前已有相同的文件被处理过，所以只需要拿之前的结果重新处理即可
	dc := outApi.DocConvert(task, task.C, "request", "")
	if task.ErrMsg != "" {
		task.ErrMsg = "图片转文档失败，详情请联系系统管理员"
		return
	}
	go base.OutApiCall(task.ToolId) // 记录api访问成功次数
	task.Result = dc.Result.TaskId
	DocConvertTask[task.Id] = 0
	go DocImgConvertGetResult(task, taskListMap)
}

func DocImgConvertGetResult(task *model.ImgTask, taskListMap map[string]bool) {
	sleepTime := 3
	ctx := context.Background()
	// 后面操作不能沿用原来的context，因为它的链接已关闭
	task.DB = g.Mysql.WithContext(ctx)
	for {
		// 每次等待一定时间
		time.Sleep(time.Duration(sleepTime) * time.Second)
		// 查询处理结果
		dc := outApi.DocConvert(task, ctx, "get_request_result", task.Result) // task.Result 此时保存的就是百度的taskID
		if task.ErrMsg != "" {                                                // 如果返回的结果有错误，则停止任务并退款
			task.Result = task.ErrMsg
			task.Update()
			// 查找订单发起退款
			order, err := model.GetOrderByTaskId(task)
			if err != nil {
				task.Log.Error("获取订单失败", err)
			}
			payService.WechatRefund(order, task.ErrMsg)
			if task.ErrMsg != "" {
				return
			}
			delete(taskListMap, task.Id) // 删除任务
			return
		}
		//查询任务进度
		DocConvertTask[task.Id] = dc.Result.Percent
		if dc.Result.Percent == 100 { // 任务成功
			delete(DocConvertTask, task.Id) // 删除map的内容，以免无限增加
			result, err := json.Marshal(dc.Result.ResultData)
			if err != nil {
				task.Log.Error("json序列化失败", err)
				task.Result = g.UnknownErrorMsg
			} else {
				task.Result = string(result)
			}
			task.Update()
			if task.ErrMsg != "" {
				return
			}
			go base.ToolCallSuccess(task.ToolId) // 记录访问次数
			delete(taskListMap, task.Id)         // 删除任务
			return
		}
		// 判断任务进行有没有超过一分钟，超过一分钟未完成退款
		if time.Now().UnixMilli()-task.CreatedAt > 60000 {
			task.Result = "图片转文档超时,任务已终止，系统已自动发起退款，其注意查收"
			delete(DocConvertTask, task.Id) // 任务超时，删除任务进度记录
			task.Update()
			// 查找订单发起退款
			order, err := model.GetOrderByTaskId(task)
			if err != nil {
				task.Log.Error("获取订单失败", err)
			}
			err = payService.WechatRefund(order, "图片转文档超时")
			if err != nil {
				task.Log.Error("发起退款失败", err)
			}
			task.Log.Warn("图片转文档超过一分钟", dc.LogId)
			delete(taskListMap, task.Id) // 删除任务记录，意味着任务已结束
			return
		}
	}
}

type DocConvertDto struct {
	Code     int            `json:"code" ` // -1：任务失败；1：进行中；2：已完成
	Percent  int            `json:"percent"`
	Result   string         `json:"result"`
	DownAddr dto.ResultData `json:"down_addr"`
}

func DocImgConvertQueryResult(task *model.ImgTask) (dc *DocConvertDto) {
	dc = &DocConvertDto{DownAddr: dto.ResultData{}}
	val, ok := DocConvertTask[task.Id]
	if ok { // 表明任务还在进行中
		dc.Code = 1
		dc.Percent = val
		return
	}
	task.FindWithCreateTime(86400)
	if task.ErrMsg != "" {
		dc.Result = task.ErrMsg
		return
	}
	// 检查结果是否是正常的json格式
	if task.Result == "" {
		dc.Code = -1
		dc.Percent = 0
		dc.Result = "未能正确获取到下载地址！"
		return
	} else if err := json.Unmarshal([]byte(task.Result), &dc.DownAddr); err != nil {
		dc.Code = -1
		dc.Percent = 0
		dc.Result = task.Result
		return
	} else if dc.DownAddr.Word == "" && dc.DownAddr.Excel == "" {
		dc.Code = -1
		dc.Percent = 0
		dc.Result = "未能正确获取到下载地址！"
		return
	}
	dc.Code = 2 // 任务已完成
	return
}
