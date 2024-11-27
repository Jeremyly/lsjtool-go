/*
 * @Author: lsjweiyi
 * @Date: 2023-08-29 19:26:39
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 12:56:48
 * @Description: 图像特效相关的实现
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package outApi

import (
	"errors"
	"net/http"
	"src/app/dto"
	"src/app/model"
	"src/global"
	"src/utils/myHttp"
)

/**
 * @description: 图像特效与增强的统一api post 请求
 * @param {context.Context} c
 * @param {[]byte} imgByte
 * @param {string} api
 * @return {*}
 */
func EffectsEnhancementPost(task *model.ImgTask, api string, errMsg string) *dto.EffectsEnhancement {
	var fullUrl string
	var sendBody *http.Request
	if task.ToolId == 26 { // 26是文档类的api
		fullUrl, sendBody = PostBaiduPrepare(task, DocumentApiUrl, api, DocumentAppId)
	} else {
		fullUrl, sendBody = PostBaiduPrepare(task, EffectsEnhancementApiUrl, api, EffectsEnhancementAppId)
	}
	if task.ErrMsg != "" {
		return nil
	}
	ee := &dto.EffectsEnhancement{}
	if err := myHttp.FormPost(task.C, fullUrl, sendBody.Form.Encode(), ee); err != nil {
		task.Log.Error(errMsg, err)
		task.ErrMsg = errMsg
		return nil
	}
	if ee.ErrorCode != 0 {
		task.Log.Error("百度图像特效与增强api返回错误", errors.New(ee.ErrorMsg))
		task.ErrMsg = global.UnknownErrorMsg
		return nil
	}
	return ee
}
