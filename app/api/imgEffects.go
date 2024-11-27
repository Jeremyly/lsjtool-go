/*
 * @Author: lsjweiyi
 * @Date: 2023-08-30 19:56:28
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 12:53:23
 * @Description:图像特效与增强相关的接口
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
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

/**
 * @description: 图像特效与增强几个通用的方法的接口，根据tooId切换所对应的功能
 * @param {*gin.Context} c
 * @param {*dto.ImgDto} bg
 * @param {int} toolId 对应的功能请看switch里的文字，
 * @return {*}
 */
func EffectsEnhancement(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	apiStr, errMsg := checkEEParams(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	service.Generic(task, apiStr, errMsg)
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

func checkEEParams(task *model.ImgTask) (string, string) {

	var apiStr, errMsg string

	maxSize := 3143680 // 限制上传大小，因为要考虑转换为base64后增大，百度接口限制4M，所以二进制数据应<4M*0.75-24
	// 从21开始，为了保留拓展性，不再延续之前的toolId，以后还有直接添加即可
	switch task.ToolId {
	case 21:
		apiStr = "image_definition_enhance"
		errMsg = "图像清晰度增强失败："
	case 22:
		apiStr = "colourize"
		errMsg = "黑白图像上色失败："
	case 23:
		apiStr = "image_quality_enhance"
		errMsg = "图像无损放大两倍失败："
	case 24:
		apiStr = "doc_repair"
		errMsg = "文档图片去底纹失败："
	case 25:
		apiStr = "remove_moire"
		errMsg = "图片去摩尔纹失败："
	case 26:
		apiStr = "remove_handwriting"
		errMsg = "图片去手写体失败："
	default:
		task.ErrMsg = "不支持的工具id"
		task.Log.Error("不支持的工具id:" + string(task.ToolId))
		return "", ""
	}
	go base.ToolCall(task.ToolId) // 记录访问工具次数

	// 因为图像的大小是百度接口要求的，而要求的是base64编码的大小，而CheckImage校验的是编码前的大小，
	// 所以传入CheckImage需要把base64MaxSize转化为源文件大小。源文件的大小约是base64编码大小的3/4。为了保证边界安全，需要再减去3个字节即24byte
	task.ImgByte, task.Suffix = service.CheckImage(task, int64(maxSize)) // 检查上传的图像的情况，确定下是什么类型的任务，限8M
	if task.ErrMsg != "" {
		return "", ""
	}

	// 核查文件类型
	if task.Suffix != "jpg" && task.Suffix != "png" && task.Suffix != "bmp" {
		task.ErrMsg = "上传的文件类型错误！"
		return "", ""
	}

	// 检查图片给尺寸
	var err error
	task.Mat, err = gocv.IMDecode(task.ImgByte, gocv.IMReadUnchanged)
	if err != nil {
		task.ErrMsg = "图像解码失败"
		return "", ""
	}
	defer task.Mat.Close()
	size := task.Mat.Size()
	if size[0] < 15 || size[1] < 15 || size[0] > 4096 || size[1] > 4096 {
		task.ErrMsg = "上传的图片尺寸不符合要求！"
		return "", ""
	}
	return apiStr, errMsg
}
