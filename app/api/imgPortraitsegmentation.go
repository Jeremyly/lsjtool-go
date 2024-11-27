/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 21:03:00
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-30 13:15:12
 * @Description:人体分析相关的接口
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package api

import (
	"encoding/json"
	"src/app/dto"
	"src/app/model"
	"src/app/service"
	g "src/global"
	"src/utils/base"
	"src/utils/response"

	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

/**
 * @description:  1.证件照换底色
 * @param {*gin.Context} c 上下文
 * @return {*}
 */
func CardChangeBGColor(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	color := checkChangeBgColorParams(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	service.ChangeBgColor(task, color)
	if task.ErrMsg != "" {
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}

	if task.Add(); task.ErrMsg != "" {
		response.Fail(c, g.CurdCreatFailCode, task.ErrMsg)
		return
	}
	go base.ToolCallSuccess(task.ToolId) // 记录访问次数
	ImgResponse(task)
}

func checkChangeBgColorParams(task *model.ImgTask) *dto.BRGA {
	if task.ToolId != 1 {
		task.ErrMsg = "接口id与路径不匹配！"
		return nil
	}
	go base.ToolCall(task.ToolId)                                 // 记录访问次数
	task.ImgByte, task.Suffix = service.CheckImage(task, 3143680) // 检查上传的图像的情况，限3070KB
	if task.ErrMsg != "" {
		return nil
	}

	if task.Suffix != "jpg" && task.Suffix != "png" && task.Suffix != "bmp" {
		task.ErrMsg = "不支持的文件格式！"
		return nil
	}
	var err error
	// 检查图片给尺寸
	task.Mat, err = gocv.IMDecode(task.ImgByte, gocv.IMReadUnchanged)
	if err != nil {
		task.ErrMsg = "图像解码失败"
		return nil
	}
	defer task.Mat.Close()
	size := task.Mat.Size()
	if size[0] < 15 || size[1] < 15 || size[0] > 4096 || size[1] > 4096 {
		task.ErrMsg = "上传的图片尺寸不符合要求！"
		return nil
	}

	var color dto.BRGA
	colorStr := task.C.PostForm("color")
	if colorStr != "" {
		if err := json.Unmarshal([]byte(colorStr), &color); err != nil {
			task.Log.Error("证件照换底色的RGB json解析失败：", err)
			task.ErrMsg = "传入的颜色不正确！"
			return nil
		}
	} else {
		return nil
	}

	return &color
}
