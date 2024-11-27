/*
 * @Author: lsjweiyi
 * @Date: 2023-07-04 21:27:11
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 12:21:44
 * @Description:
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package outApi

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"src/app/model"
	g "src/global"
)

/**
 * @description: 请求接口前的准备工作，到这一步仍未开始请求的，完成URL拼接和参数的编码
 * @param {context.Context} c
 * @param {[]byte} imgByte
 * @param {string} url 外部api 的前段统一地址
 * @param {string} api 不同api接口对应的接口名称
 * @param {string} appId 对应的应用id
 * @return {*} 拼接后完整的url；请求的参数结构体(返回出去是为了不同接口的定制化参数的添加)；错误信息
 */
func PostBaiduPrepare(task *model.ImgTask, url string, api string, appId int) (fullUrl string, sendBody *http.Request) {
	host := url + api + "?access_token=%s"

	// 请求参数拼接
	thisSendBody := http.Request{}
	thisSendBody.ParseForm()
	thisSendBody.Form.Add("image", base64.StdEncoding.EncodeToString(task.ImgByte))
	fullUrl = fmt.Sprintf(host, g.BaiduToken[appId])
	sendBody = &thisSendBody
	return
}
