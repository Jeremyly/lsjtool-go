/*
 * @Author: lsjweiyi
 * @Date: 2023-09-06 18:54:17
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 15:16:01
 * @Description:用于抽离一些重复代码
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package api

import (
	"fmt"
	"path"
	"src/app/model"
	"src/app/service"
	g "src/global"
	"src/utils/base"
	"src/utils/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

/**
 * @description: 图片下载
 * @param {*gin.Context} c
 * @return {*}
 */
func ImageDown(c *gin.Context) {

	toolId, err := strconv.Atoi(c.Param("toolId"))
	if err != nil {
		response.Fail(c, g.ValidatorParamsCheckFailCode, "工具id输入错误")
		return
	}
	_, ok := model.ToolsMap[uint8(toolId)]
	if !ok {
		response.Fail(c, g.ValidatorParamsCheckFailCode, "工具不存在")
		return
	}
	id := c.Param("taskId")
	task := &model.ImgTask{Id: id, ToolId: uint8(toolId), BaseModel: *model.InitBaseModel(c).GetCtx(c, c).GetLog(c).GetDB(c, c)}

	if task.ToolId == 27 {
		response.Fail(c, g.ValidatorParamsCheckFailCode, "该类型的资源请到对应页面中进行下载")
		return
	}
	isReduceDownTimes := true
	// 为了适配H5端有些浏览器不能很好的支持blob的下载方式，是能让其用a标签下载
	// 此外，手机端点击下载后，会发起两次请求，导致下载次数增加，所以这里对手机端进行特殊处理
	if c.Request.Header.Get("Sec-Fetch-Dest") == "document" { // 该标记就是判断手机端的第一次发起请求，第一次就不减少下载次数
		isReduceDownTimes = false
	}
	service.DownImg(task, isReduceDownTimes)
	if task.ErrMsg != "" {
		response.Fail(c, g.ExecuteErrorCode, task.ErrMsg)
		return
	}

	filePath := path.Join(model.ToolsMap[task.ToolId].SaveDir, fmt.Sprintf(`result_%s.%s`, task.Id, task.Suffix))
	c.Writer.Header().Add("Access-Control-Expose-Headers", "Content-Disposition,down-times") // 暴露该请求头给前端
	c.Header("down-times", strconv.Itoa(task.DownloadTimes))

	go base.ToolDownload(task.ToolId)
	c.FileAttachment(filePath, fmt.Sprintf(`%s.%s`, task.Id, task.Suffix))
}

func ImgResponse(task *model.ImgTask) {
	task.C.Writer.Header().Add("Access-Control-Expose-Headers", "Content-Disposition") // 暴露该请求头给前端
	task.C.FileAttachment(task.ResultPath, fmt.Sprintf(`%s.%s`, task.Id, task.Suffix))
}

// CheckSuffix 函数用于检查给定的后缀是否在可接受的后缀列表中
// 参数suffix为待检查的后缀，参数suffixAccept为可接受的后缀列表
// 若suffix在suffixAccept中，则返回true，否则返回false
func CheckSuffix(suffix string, suffixAccept []string) bool {
	for _, v := range suffixAccept {
		if v == suffix {
			return true
		}
	}
	return false
}

// GetToolMsg 用于更新工具信息的接口
func GetToolMsg(c *gin.Context) {
	toolId, err := strconv.Atoi(c.Request.Header.Get(g.ToolIdKey))
	if err != nil {
		response.Fail(c, g.ExecuteErrorCode, "工具id输入错误")
		return
	}
	tool, ok := model.ToolsMap[uint8(toolId)]
	if !ok {
		response.Fail(c, g.ExecuteErrorCode, "工具不存在")
		return
	}
	response.Success(c, tool)

}

// 获取taskId对应的下载次数
func GetDownTimes(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	task.FindWithCreateTime(86400) // 只允许查找一天内的数据
	if task.ErrMsg != "" {
		response.Fail(c, g.CurdSelectFailCode, task.ErrMsg)
		return
	}
	response.Success(c, task.DownloadTimes)
}

func UpdateToolsMsg(c *gin.Context) {
	if err := model.LoadToolConfig(); err != nil {
		response.Fail(c, g.ExecuteErrorCode, "更新工具信息失败")
		return
	}
	response.Success(c, "更新工具信息成功")
}
