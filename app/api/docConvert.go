/*
 * @Author: lsjweiyi
 * @Date: 2024-05-02 19:38:12
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-02 22:05:13
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package api

import (
	"context"
	"errors"
	"net/http"
	"src/app/model"
	"src/app/service"
	"src/app/service/payService"
	g "src/global"
	"src/utils/base"
	"src/utils/response"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

var (
	taskChannel = make(chan uint8, 1)   // 防止并发的通道
	taskListMap = make(map[string]bool) // 记录任务是否在进行中
)

/**
 * @description: 文档转换
 * @param {*gin.Context} c
 * @return {*}
 */
func DocImgConvert(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	if err := queryTaskStatus(task); err != nil { // 这里面有调用taskChannel
		response.Fail(c, g.ExecuteErrorCode, err)
		return
	}

	order, err := model.GetOrderByTaskId(task)
	if err != nil {
		delete(taskListMap, task.Id) // 删除任务
		response.Fail(c, g.CurdSelectFailCode, err)
		return
	}
	if order.ResultType != 0 {
		delete(taskListMap, task.Id) // 删除任务
		response.Fail(c, g.ExecuteErrorCode, "该工具需要先付款再使用")
		return
	}
	task.UpdatedAt = time.Now().UnixMilli() // 这里相当于记录任务开始的时间
	checkDCImgParams(task)
	if task.ErrMsg != "" {
		delete(taskListMap, task.Id) // 删除任务
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	service.DocImgConvertRequest(task, taskListMap)
	if task.ErrMsg != "" {
		delete(taskListMap, task.Id) // 删除任务
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}
	response.Success(c, "文档转换已开始，该任务需要一点时间，请稍等...")
}

/**
 * @description: 查询文档转换结果
 * @param {*gin.Context} c
 * @return {*}
 */
func DocConvertResult(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	dc := service.DocImgConvertQueryResult(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}
	if dc.Code == -1 {
		response.Fail(c, dc.Code, dc.Result)
		return
	}
	docT := docType{}
	if dc.Code == 2 {
		if dc.DownAddr.Word != "" {
			docT.Word = true
		}
		if dc.DownAddr.Excel != "" {
			docT.Excel = true
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": dc.Code,
		"data": gin.H{"message": dc.Result, "word": docT.Word, "excel": docT.Excel},
	})
}

type docType struct {
	Word  bool `json:"word"`
	Excel bool `json:"excel"`
}

var (
	downLock = sync.Mutex{}
)

// 获取对应类型的下载链接
func Get27ToolDownUrl(c *gin.Context) {
	docTypeStr := c.Param("type")
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	downLock.Lock() // 加锁，防止并发
	defer downLock.Unlock()
	dc := service.DocImgConvertQueryResult(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}
	if dc.Code != 2 {
		response.Fail(c, dc.Code, dc.Result)
		return
	}
	if task.DownloadTimes <= 0 {
		response.Fail(c, g.ExecuteErrorCode, "下载次数已用完")
		return
	}
	if docTypeStr == "word" {
		go base.ToolDownload(task.ToolId)
		task.ReduceDownloadTimes()                                           // 减少下载次数
		c.Writer.Header().Add("Access-Control-Expose-Headers", "down-times") // 暴露该请求头给前端
		c.Header("down-times", strconv.Itoa(task.DownloadTimes))
		response.Success(c, dc.DownAddr.Word)
	} else if docTypeStr == "excel" {
		go base.ToolDownload(task.ToolId)
		task.ReduceDownloadTimes()                                           // 减少下载次数
		c.Writer.Header().Add("Access-Control-Expose-Headers", "down-times") // 暴露该请求头给前端
		c.Header("down-times", strconv.Itoa(task.DownloadTimes))
		response.Success(c, dc.DownAddr.Excel)
	} else {
		response.Fail(c, g.ExecuteErrorCode, "不支持的类型")
		return
	}
}

// 检查参数
func checkDCImgParams(task *model.ImgTask) {
	if task.ToolId != 27 {
		task.ErrMsg = "接口id与路径不匹配！"
		return
	}
	go base.ToolCall(task.ToolId)                                 // 记录访问次数
	task.ImgByte, task.Suffix = service.CheckImage(task, 3143680) // 检查上传的图像的情况，限3070KB
	if task.ErrMsg != "" {
		return
	}

	if task.Suffix != "jpg" && task.Suffix != "png" && task.Suffix != "bmp" {
		task.ErrMsg = "不支持的文件格式！"
		return
	}

	// 检查图片给尺寸
	mat, err := gocv.IMDecode(task.ImgByte, gocv.IMReadUnchanged)
	if err != nil {
		task.ErrMsg = "图像解码失败"
		return
	}
	defer mat.Close()
	size := mat.Size()
	if size[0] < 15 || size[1] < 15 || size[0] > 4096 || size[1] > 4096 {
		task.ErrMsg = "上传的图片尺寸不符合要求！"
		return
	}

}

// 查询task
func queryTaskStatus(task *model.ImgTask) error {
	taskChannel <- 1 // 阻塞，防止并发
	_, ok := taskListMap[task.Id]
	if ok {
		<-taskChannel // 释放
		return errors.New("该任务正在执行中，请勿重复执行")
	}
	// 查找数据库记录，确认该任务是否已经执行完成
	if task.Find(); task.ErrMsg != "" {
		<-taskChannel // 释放
		return errors.New(task.ErrMsg)
	}
	if task.Result != "" {
		<-taskChannel // 释放
		return errors.New("该任务已经执行完成，请勿重复执行")
	}
	taskListMap[task.Id] = true // 标记任务正在执行
	<-taskChannel               // 释放
	return nil
}

// 针对27号工具，查询已支付超过10分钟但未使用的订单，自动发起退款
func QueryOrderAutoRefund27() {
	thisLog := g.Log{RequestUrl: "QueryOrderAutoRefund27"}
	// 先从数据库中查询支付状态为0且订单创建时间超过10分钟的订单，然后主动发起订单查询
	var orders []model.Order
	err := g.Mysql.Table("order").Where("tool_id = ? and result_type = ? and end_at < ?", 27, 0, time.Now().Add(-10*time.Minute).UnixMilli()).Find(&orders).Error
	if err != nil {
		thisLog.Error("查询订单失败", err)
		return
	}
	ctx := context.Background()
	// 发起订单查询
	for _, order := range orders {
		task := model.ImgTask{Id: order.TaskId, ToolId: order.ToolId}
		// task需要初始化
		task.Ctx = ctx
		task.DB = g.Mysql
		task.Log = &thisLog

		// 查询任务状态是否满足退款要求
		taskChannel <- 1 // 阻塞，防止并发
		_, ok := taskListMap[task.Id]
		if ok {
			<-taskChannel // 释放
			continue
		}
		// 查找数据库记录，确认该任务是否已经执行完成
		if task.Find(); task.ErrMsg != "" {
			<-taskChannel // 释放
			thisLog.Error("在自动退单任务中查询任务失败:"+task.Id, errors.New(task.ErrMsg))
			continue
		}
		if task.Result != "" {
			<-taskChannel // 释放
			continue
		}
		<-taskChannel // 释放
		// 订单也需要初始化
		order.DB = g.Mysql
		order.Log = &thisLog
		order.Ctx = ctx
		if order.Platform == 1 {
			if err := payService.WechatRefund(&order, "超时未使用自动退款"); err != nil {
				thisLog.Error("自动退款失败:"+order.Id, err)
			}
		} else {
			if err := payService.AliRefund(&order, "超时未使用自动退款"); err != nil {
				thisLog.Error("自动退款失败:"+order.Id, err)
			}
		}
	}
}
