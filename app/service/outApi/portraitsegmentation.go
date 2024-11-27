/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:48:43
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-17 14:20:24
 * @Description:对接百度人体分析的接口
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */

package outApi

import (
	"errors"
	"src/app/dto"
	"src/app/model"
	"src/global"
	"src/utils/myHttp"
)

/*
*
  - @description: 百度人像分割Api
  - @param {*gin.Context} c 上下文
  - @param {string} imgBase64 图像的base64编码
  - @param {string} resType 返回值包含的内容。type 参数值可以是可选值的组合，用逗号分隔；如果无此参数默认输出全部3类结果图。可选值说明：
    labelmap - 二值图像，需二次处理方能查看分割效果
    scoremap - 人像前景灰度图
    foreground - 人像前景抠图，透明背景。
  - @return {*dto.PortraitResp} 解析后的结构体
*/
func PortraitSegmentation(task *model.ImgTask, resType string) *dto.PortraitResp {
	fullUrl, sendBody:= PostBaiduPrepare(task, PortraitApiUrl, "body_seg", PortraitAppId)
	if task.ErrMsg != "" {
		return nil
	}
	if resType != "" {
		sendBody.Form.Add("type", resType)
	}
	var pr dto.PortraitResp
	if err := myHttp.FormPost(task.Ctx, fullUrl, sendBody.Form.Encode(), &pr); err != nil {
		task.Log.Error("请求百度人像分割api失败", err)
		task.ErrMsg = global.UnknownErrorMsg
		return nil
	}
	if pr.ErrorCode != 0 {
		task.Log.Error("百度人像分割api返回错误", errors.New(pr.ErrorMsg))
		task.ErrMsg = global.UnknownErrorMsg
		return nil
	}
	return &pr
}
