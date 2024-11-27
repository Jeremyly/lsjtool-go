/*
 * @Author: error: git config user.name & please set dead value or install git
 * @Date: 2023-08-07 18:31:09
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-04 01:16:29
 * @Description: 图像质量相关的api
 * Copyright (c) 2023 by error: git config user.name & please set dead value or install git, All Rights Reserved.
 */
package api

import (
	"fmt"
	"os"
	"src/app/model"
	"src/app/service"
	g "src/global"
	"src/utils/base"
	"src/utils/response"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

/**
 * @description: png 无损压缩的api接口
 * @param {*gin.Context} c
 * @return {*}
 */
func PngLosslessCompress(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	compressLevel := checkPngLosslessCompressParams(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	service.PngLosslessCompress(task, compressLevel)
	if task.ErrMsg != "" {
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}

	if task.Add(); task.ErrMsg != "" {
		response.Fail(c, g.CurdCreatFailCode, task.ErrMsg)
		return
	}
	go base.ToolCallSuccess(task.ToolId) // 记录访问次数

	fs, err := os.Stat(task.ResultPath)
	if err != nil {
		task.Log.Error("获取文件信息出错：", err)
		response.Fail(c, g.UnknownErrorCode, "")
		return
	}
	response.Success(c, fs.Size())
}

/**
 * @description: 图片质量压缩API
 * @param {*gin.Context} c
 * @return {*} 无返回，该任务发起后不等待结果出来，而是直接返回，由CheckimgQuantityTask后续查询结果
 */
func LossyCompression(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	needSize := checkLossyCompressParams(task)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	service.LossyCompression(task, needSize)
	if task.ErrMsg != "" {
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}

	fs, err := os.Stat(strings.Replace(task.ResultPath, "water", "result", 1))
	if err != nil {
		task.Log.Error("获取文件信息出错：", err)
		response.Fail(c, g.UnknownErrorCode, g.UnknownErrorMsg)
		return
	}
	if task.Add(); task.ErrMsg != "" {
		response.Fail(c, g.CurdCreatFailCode, task.ErrMsg)
		return
	}
	go base.ToolCallSuccess(task.ToolId) // 记录成功访问次数

	c.Writer.Header().Add("Access-Control-Expose-Headers", "Compress-Size")       // 暴露该请求头给前端
	c.Writer.Header().Add("Access-Control-Expose-Headers", "Content-Disposition") // 暴露该请求头给前端
	c.Header("Compress-Size", strconv.FormatInt(fs.Size(), 10))                   // 设置response的header
	c.FileAttachment(task.ResultPath, fmt.Sprintf(`%s.%s`, task.Id, task.Suffix))
}

/**
 * @description: 有损压缩的参数校验
 * @param {*model.ImgTask} task
 * @return {*}
 */
func checkLossyCompressParams(task *model.ImgTask) int {
	if task.ToolId != 2 {
		task.ErrMsg = "接口id与路径不匹配！"
		return 0
	}
	go base.ToolCall(task.ToolId)                                  // 记录访问次数
	task.ImgByte, task.Suffix = service.CheckImage(task, 10485760) // 检查上传的图像的情况，确定下是什么类型的任务，限10M
	if task.ErrMsg != "" {
		return 0
	}
	// 支持的格式jpg png
	suffixAccept := []string{"jpg", "jpeg", "png"}
	// 检查 传入的文件类型和需要的文件类型是否满足suffixAccept的要求
	if !CheckSuffix(task.Suffix, suffixAccept) {
		task.ErrMsg = "上传的文件不是支持的格式!"
		return 0
	}

	needSizeStr := task.C.PostForm("needSize")
	needSize, err := strconv.Atoi(needSizeStr)
	if err != nil {
		task.ErrMsg = "输入的目标大小不是整数"
		return 0
	}
	if needSize <= 0 {
		task.ErrMsg = "输入的目标大小不得小于0"
		return 0
	}
	return needSize
}

/**
 * @description: png 无损压缩的参数校验
 * @param {*model.ImgTask} task
 * @return {*}
 */
func checkPngLosslessCompressParams(task *model.ImgTask) int {
	if task.ToolId != 3 {
		task.ErrMsg = "接口id与路径不匹配！"
		return 0
	}
	go base.ToolCall(task.ToolId)                                  // 记录访问次数
	task.ImgByte, task.Suffix = service.CheckImage(task, 10485760) // 检查上传的图像的情况，确定下是什么类型的任务，限10M
	if task.ErrMsg != "" {
		return 0
	}

	// 核查文件类型
	if task.Suffix != "png" {
		task.ErrMsg = "上传的文件类型错误！"
		return 0
	}

	compressLevelStr := task.C.PostForm("level")
	compressLevel, err := strconv.Atoi(compressLevelStr)
	if err != nil {
		task.ErrMsg = "输入的压缩等级不是整数"
		return 0
	}
	if !(compressLevel >= 0 && compressLevel < 10) {
		task.ErrMsg = "输入的压缩等级超出范围"
		return 0
	}
	return compressLevel
}
