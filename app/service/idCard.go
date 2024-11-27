/*
 * @Author: lsjweiyi
 * @Date: 2023-06-29 20:42:39
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-26 21:01:40
 * @Description: 证件照处理的相关实现
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */

package service

import (
	"src/app/dto"
	"src/app/model"
	"src/app/service/outApi"
	g "src/global"
	"src/utils/base"

	"gocv.io/x/gocv"
)

/**
 * @description: 证件照换底色
 * @param {*dto.BgColorBody} bg 原图片的信息
 * @return {*} 处理好的证件照
 */
func ChangeBgColor(task *model.ImgTask, color *dto.BRGA) {

	fg := outApi.PortraitSegmentation(task, "foreground")
	if task.ErrMsg != "" {
		return
	}
	go base.OutApiCall(task.ToolId) // 记录api访问次数
	var err error
	task.Mat, err = ImgStr2RGB(fg.Foreground)
	if err != nil {
		task.Log.Error("证件照换底色失败：", err)
		task.ErrMsg = g.UnknownErrorMsg
		return
	}
	defer task.Mat.Close()
	// 如果传入的背景色为空，则不进行换底色操作，直接返回分割结果
	if color == nil {
		// 保存原始结果图
		task.Suffix = "png" // 原图要保持透明，就必须是png格式
		if GocvSaveImg(task, "result_", -1); task.ErrMsg != "" {
			return
		}
		ImgWatermark(task.Mat) // 给图片添加水印
		// 保存带水印的结果图
		if task.ResultPath = GocvSaveImg(task, "water_", 80); task.ErrMsg != "" {
			return
		}
	} else {
		cardChangeBackground(task, color)
	}
}

/**
 * @description: 通过接口分割完人像后，将背景替换成需要的颜色
 * @param {gocv.Mat} foreGMat 分割完后的人像图
 * @param {*gocv.Scalar} bgScalar 目的背景色
 * @return {*} 处理后的照片转为base64
 */
func cardChangeBackground(task *model.ImgTask, color *dto.BRGA) {
	// 转化为float，且把值转化为[0-1]
	task.Mat.ConvertToWithParams(&task.Mat, gocv.MatTypeCV32FC1, 1.0/255.0, 0)
	//分通道
	spiltImg := gocv.Split(task.Mat)

	//创建背景矩阵且把值转化为[0-1]
	bgMat := gocv.NewMatWithSizeFromScalar(*color.GetOneScalar(), task.Mat.Rows(), task.Mat.Cols(), gocv.MatTypeCV32FC3)

	// 将Alpha通道拼接成3通道，和其他矩阵保持一样的通道数，以便计算
	d3Mat := gocv.NewMat()
	gocv.Merge([]gocv.Mat{spiltImg[3], spiltImg[3], spiltImg[3]}, &d3Mat)

	// 将前三个通道取出，方便计算
	img3 := gocv.NewMat()
	gocv.Merge([]gocv.Mat{spiltImg[0], spiltImg[1], spiltImg[2]}, &img3)
	// 用完马上关闭，防止内存占用
	for _, v := range spiltImg {
		v.Close()
	}

	// 这里将公式 C=d*A+(1-d)*B转换为C=d*(A-B)+B,下面是拆解逐步计算
	// A-B
	aSubB := gocv.NewMat()
	gocv.Subtract(img3, bgMat, &aSubB)
	img3.Close()

	// d*(A-B)
	dMulAB := gocv.NewMat()
	gocv.Multiply(d3Mat, aSubB, &dMulAB)
	d3Mat.Close()
	aSubB.Close()

	task.Mat.Close() // 这里需要先关闭，后面重新赋值了

	// d*(A-B)+B
	task.Mat = gocv.NewMat()

	gocv.Add(dMulAB, bgMat, &task.Mat)
	bgMat.Close()
	dMulAB.Close()
	// 转化为正常的[0-255]图片
	task.Mat.ConvertToWithParams(&task.Mat, gocv.MatTypeCV8UC1, 255, 0)
	// 保存原始结果图
	if GocvSaveImg(task, "result_", -1); task.ErrMsg != "" {
		return
	}
	ImgWatermark(task.Mat) // 给图片添加水印
	// 保存带水印的结果图
	if task.ResultPath = GocvSaveImg(task, "water_", 80); task.ErrMsg != "" {
		return
	}
}
