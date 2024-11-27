/*
 * @Author: lsjweiyi
 * @Date: 2023-08-29 19:47:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-25 20:41:47
 * @Description:图像特效与增强几个调用方式通用的api的实现
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package service

import (
	"src/app/model"
	"src/app/service/outApi"
	g "src/global"
	"src/utils/base"
)

/**
 * @description: 图像特效与增强几个调用方式通用的api的实现,会保存处理后的结果图，也会保存处理后的水印图
 * @param {*model.ImgTask} task
 * @param {string} api 各api对应的百度的接口路径的最后不同的那部分
 * @param {string} errMsg 错误信息，为了让各接口返回不同的错误信息
 */
func Generic(task *model.ImgTask, api string, errMsg string) {
	// 之前已有相同的文件被处理过，所以只需要拿之前的结果重新处理即可
	ee := outApi.EffectsEnhancementPost(task, api, errMsg)
	if task.ErrMsg != "" {
		return
	}
	go base.OutApiCall(task.ToolId) // 记录api访问成功次数
	var err error
	task.Mat, err = ImgStr2RGB(ee.GetImage())
	if err != nil {
		task.Log.Error(errMsg, err)
		task.ErrMsg = g.UnknownErrorMsg
		return
	}
	if GocvSaveImg(task, "result_", -1); task.ErrMsg != "" {
		return
	}
	// 加水印后保存水印照
	ImgWatermark(task.Mat)
	if task.ResultPath = GocvSaveImg(task, "water_", -1); task.ErrMsg != "" {
		return
	}
	task.Mat.Close()
}
