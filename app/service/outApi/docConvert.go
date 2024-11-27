/*
 * @Author: lsjweiyi
 * @Date: 2024-05-02 16:30:22
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-02 21:40:30
 * @Description: 调用百度文档格式转换api
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package outApi

import (
	"context"
	"errors"
	"net/http"
	"src/app/dto"
	"src/app/model"
	"src/utils/myHttp"
)

/**
 * @description: 文档格式转换
 * @param {*model.ImgTask} task
 * @param {string} api
 * @param {string} taskId
 * @return {*}
 */
func DocConvert(task *model.ImgTask, ctx context.Context, api string, taskId string) *dto.DocConvert {
	var fullUrl string
	var sendBody *http.Request
	fullUrl, sendBody = PostBaiduPrepare(task, DocumentApiUrl, "doc_convert/"+api, DocumentAppId)

	if task.ErrMsg != "" {
		return nil
	}
	// 获取结果时，需要上传task_id
	if api == "get_request_result" {
		sendBody.Form.Add("task_id", taskId)
	}

	dc := &dto.DocConvert{}
	if err := myHttp.FormPost(task.C, fullUrl, sendBody.Form.Encode(), dc); err != nil {
		task.Log.Error("百度文档格式转换api请求失败", err)
		task.ErrMsg = err.Error()
		return nil
	}
	if dc.ErrorCode != 0 {
		task.Log.Error("百度文档格式转换api返回错误", errors.New(dc.ErrorMsg))
		task.ErrMsg = dc.ErrorMsg
		return nil
	}
	if dc.Code != 1001 {
		task.Log.Error("百度文档格式转换api返回错误", errors.New(dc.Message))
		task.ErrMsg = dc.Message
		return nil
	}

	return dc
}
