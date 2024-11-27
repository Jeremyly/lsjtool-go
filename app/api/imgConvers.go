/*
 * @Author: lsjweiyi
 * @Date: 2024-03-24 11:14:32
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-03 22:48:26
 * @Description: 图像格式转换的接口
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package api

import (
	"src/app/model"
	"src/app/service"
	g "src/global"
	"src/utils/base"
	"src/utils/response"

	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

// 基于 gocv 实现图片格式转换
func ImgConvers(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	checkSuffixParams(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	var err error
	task.Mat, err = gocv.IMDecode(task.ImgByte, gocv.IMReadUnchanged)
	if err != nil {
		task.Log.Error("解析图片失败：", err)
		response.Fail(c, g.ValidatorParamsCheckFailCode, "解析图片失败")
		return
	}
	defer task.Mat.Close()
	service.GocvSaveImg(task, "result_", -1)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	service.ImgWatermark(task.Mat)
	task.ResultPath = service.GocvSaveImg(task, "water_", -1)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	task.Add()
	if task.ErrMsg != "" {
		response.Fail(c, g.CurdCreatFailCode, task.ErrMsg)
		return
	}
	go base.ToolCallSuccess(task.ToolId) // 记录访问次数
	response.Success(c, task.Suffix)
}

// 检查上传的图片格式是否正确
func checkSuffixParams(task *model.ImgTask) {
	if task.ToolId != 4 {
		task.ErrMsg = "接口id与路径不匹配！"
		return
	}
	go base.ToolCall(task.ToolId)                                       // 记录访问次数
	task.ImgByte, task.ReginSuffix = service.CheckImage(task, 10485760) // 检查上传的图像的情况，确定下是什么类型的任务，限10M
	if task.ErrMsg != "" {
		return
	}
	task.Suffix = task.C.PostForm("suffix")
	if task.ReginSuffix == task.Suffix {
		task.ErrMsg = "上传的文件和转换格式相同,无需转换"
		return
	}
	// 支持的格式jpg jp2 png tif bmp dwg
	suffixAccept := []string{"jpg", "jpeg", "jp2", "png", "tiff", "bmp", "webp"}

	// 检查 传入的文件类型和需要的文件类型是否满足suffixAccept的要求
	if !CheckSuffix(task.ReginSuffix, suffixAccept) {
		task.ErrMsg = "上传的文件不是支持的格式!"
		return
	}
	if !CheckSuffix(task.Suffix, suffixAccept) {
		task.ErrMsg = "所需的转换格式不支持!"
		return
	}
}
