/*
 * @Author: lsjweiyi
 * @Date: 2023-07-05 19:33:06
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-02 21:05:39
 * @Description:
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package myHttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

/**
 * @description: 发起post请求
 * @param {string} urlStr 请求的url
 * @param {string} paramStr 请求的参数拼接成的字符串，形如：id=1&secret=sa&type=a
 * @param {any} jsonBody 返回数据对应的结构体
 * @return {*} 错误信息
 */
func FormPost(c context.Context, urlStr string, paramStr string, jsonBody any) error {

	payload := strings.NewReader(paramStr)
	client := &http.Client{}
	// 使用带上下文的请求，当这个请求时间太长时，会被gin的上下文中断。
	req, err := http.NewRequestWithContext(c, http.MethodPost, urlStr, payload)
	if err != nil {
		return fmt.Errorf("输入的url或参数异常:%s", err.Error())
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败:%s", err.Error())
	}
	defer response.Body.Close()
	// 将返回的内容全部读取
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("读取返回数据失败:%s", err.Error())
	}
	// 对内容进行解析
	err = json.Unmarshal(result, jsonBody)
	if err != nil {
		return fmt.Errorf("Json解析失败:%s", err.Error())
	}
	return nil
}
