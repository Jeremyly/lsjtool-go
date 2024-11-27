/*
 * @Author: lsjweiyi
 * @Date: 2023-08-29 19:39:10
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-02 22:11:21
 * @Description:外部api返回数据的json 结构体
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package dto

type ThisError struct {
	ErrorCode int32  `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

/**
 * @description: 图像效果增强的结构体，用于外部api返回的信息的json解析
 * @return {*}
 */
type EffectsEnhancement struct {
	Image          string
	Result         string
	ImageProcessed string `json:"image_processed"`
	LogId          uint64 `json:"log_id"`
	ThisError
}

/**
 * @description: 由于多个api接口返回的字段名不一致，所以需要判断到底哪个不为空，哪个不为空就返回哪个。默认是image
 * @return {*}
 */
func (e *EffectsEnhancement) GetImage() string {
	if e.Image != "" {
		return e.Image
	}
	if e.Result != "" {
		return e.Result
	}
	if e.ImageProcessed != "" {
		return e.ImageProcessed
	}
	return e.Image
}

/**
 * @description: 人像分割的返回值结构体
 */
type PortraitResp struct {
	PersonNum  int32 `json:"person_num"`
	Foreground string
	Scoremap   string
	Labelmap   string
	ThisError
}

/**
 * @description: 文档转换的返回值结构体
 * @return {*}
 */
type DocConvert struct {
	Code    int    `json:"code"`
	LogId   uint64 `json:"log_id"`
	Message string `json:"message"`
	Result  struct {
		TaskId     string     `json:"task_id"`
		RetCode    int        `json:"ret_code"`
		Percent    int        `json:"percent"`
		ResultData ResultData `json:"result_data"`
	} `json:"result"`
	Success bool `json:"success"`
	ThisError
}

type ResultData struct {
	Word  string `json:"word"`
	Excel string `json:"excel"`
}
