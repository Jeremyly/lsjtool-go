/*
 * @Author: error: git config user.name & please set dead value or install git
 * @Date: 2023-08-05 09:31:53
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-12 11:59:35
 * @Description:图像质量相关的实现
 * Copyright (c) 2023 by error: git config user.name & please set dead value or install git, All Rights Reserved.
 */
package service

import (
	"errors"
	"image/png"
	"src/app/model"
	g "src/global"
	myFileUtil "src/utils/file"

	"github.com/ultimate-guitar/go-imagequant"
	"gocv.io/x/gocv"
)

const (
	pngUnknownErrorMsg string = "图片进行压缩时发生未知异常，请联系管理员或稍后再试"
)

type compressBestWay struct {
	WayType  int // 1:使用pngquant进行32位转8位。2:JPG压缩
	WayParam int // 1:对应的是图像压缩等级, 2:对应jpg等较常规的压缩，压缩质量的参数，通常是0-100
}

/**
 * @description: 图像有损压缩
 * @param {*model.ImgTask} 任务对象
 * @return {*}无
 */
func LossyCompression(task *model.ImgTask, needSize int) {
	var err error
	task.Mat, err = gocv.IMDecode(task.ImgByte, gocv.IMReadUnchanged)
	if err != nil {
		task.Log.Error("解析图片失败：", err)
		task.ErrMsg = "解析图片失败"
		return
	}
	defer task.Mat.Close()
	var compressWay *compressBestWay

	// 目前OpenCV仅支持这两种常见格式的压缩保存，其他格式不常见
	switch task.Suffix {
	case "jpg":
		compressWay = jpgCompress(task, needSize)
	case "png":
		compressWay = pngCompress(task, needSize)
	}
	// 总结上述操作的结果
	if task.ErrMsg != "" {
		return
	}

	ImgWatermark(task.Mat)                                          // 添加水印
	task.ResultPath = myFileUtil.GetImgSaveFullPath(task, "water_") // 结果保存路径
	if err = comfirmCompress(task, compressWay); err != nil {       //随后以确定的方案处理处结果并保存
		task.ErrMsg = err.Error()
	}
}

/**
 * @description: jpg格式图片的压缩
 * @param {*model.ImgTask} 图像对象
 * @param {int} needSize 需要压缩到的目标大小
 * @return {*} 最优压缩方式
 */
func jpgCompress(task *model.ImgTask, needSize int) *compressBestWay {
	cvCompressParam := []int{gocv.IMWriteJpegQuality, 100} //开始不压缩，因为如果是其他格式转换过来的，不需要压缩都大概率满足条件
	for {
		imgbuf, thisErr := gocv.IMEncodeWithParams(gocv.JPEGFileExt, task.Mat, cvCompressParam)
		if thisErr != nil {
			task.Log.Error("jpg使用gocv压缩时失败", thisErr)
			task.ErrMsg = g.UnknownErrorMsg
			return nil
		}
		if imgbuf.Len() <= needSize {
			myFileUtil.SaveImg(task, imgbuf.GetBytes(), "result_")
			if task.ErrMsg != "" {
				return nil
			}
			return &compressBestWay{WayType: 2, WayParam: cvCompressParam[1]}
		}
		if cvCompressParam[1] <= 0 { // 压到底了，也正常保存了
			myFileUtil.SaveImg(task, imgbuf.GetBytes(), "result_")
			if task.ErrMsg != "" {
				return nil
			}
			return &compressBestWay{WayType: 2, WayParam: cvCompressParam[1]}
		}
		cvCompressParam[1] -= 5
		imgbuf.Close()
	}
}

/**
 * @description: 依赖pngquant 实现png有损压缩，pngquant实现的的32位转8位，CompressionLevel是png无损压缩等级。
 * @param {*model.ImgTask} 图像信息
 * @param {int} needSize 用户所需的压缩后的大小，值小于它即可，单位byte
 * @return {*} 最优的压缩方法；错误
 */
func pngCompress(task *model.ImgTask, needSize int) (pcb *compressBestWay) {
	// 首先是采用32位转8位，如果小于目标大小，则直接保存
	outByte, err := imagequant.Crush(task.ImgByte, 10, png.NoCompression)
	if err != nil {
		task.Log.Error("使用pngquant压缩时失败", err)
		task.ErrMsg = g.UnknownErrorMsg
		return
	}
	if len(outByte) < needSize {
		myFileUtil.SaveImg(task, outByte, "result_")
		if task.ErrMsg != "" {
			return nil
		}
		pcb = &compressBestWay{WayType: 1, WayParam: -1}
		return
	}
	// 其次采用pngquant的速度最快的压缩法，如果小于目标大小，则直接保存
	outByte, err = imagequant.Crush(task.ImgByte, 10, png.BestSpeed)
	if err != nil {
		task.Log.Error("使用pngquant压缩时失败", err)
		task.ErrMsg = g.UnknownErrorMsg
		return
	}
	if len(outByte) < needSize {
		myFileUtil.SaveImg(task, outByte, "result_")
		if task.ErrMsg != "" {
			return nil
		}
		pcb = &compressBestWay{WayType: 1, WayParam: -2}
		return
	}
	// 最后采用最高压缩质量，如果小于目标大小，则直接保存
	outByte, err = imagequant.Crush(task.ImgByte, 10, png.BestCompression)
	if err != nil {
		task.Log.Error("使用pngquant压缩时失败", err)
		task.ErrMsg = g.UnknownErrorMsg
		return
	}
	// 不管是否达到目标，都直接保存
	myFileUtil.SaveImg(task, outByte, "result_")
	if task.ErrMsg != "" {
		return nil
	}
	pcb = &compressBestWay{WayType: 1, WayParam: -3}
	return
}

/**
 * @description: png图像在已有明确的压缩参数下进行压缩，依赖pngQuantCompress等方法返回
 * @param {*model.ImgTask} 图像信息
 * @param {*PngCompressBestWay} compressWay压缩参数
 * @param {string} saveFile 保存的路径
 * @return {*}
 */
func comfirmCompress(task *model.ImgTask, compressWay *compressBestWay) (err error) {
	err = errors.New(pngUnknownErrorMsg)
	switch compressWay.WayType {
	case 1: //pngquant压缩
		cvCompressParam := []int{gocv.IMWritePngCompression, 0, gocv.IMWritePngStrategy, 0}
		imgbuf, thisErr := gocv.IMEncodeWithParams(".png", task.Mat, cvCompressParam)
		if thisErr != nil {
			task.Log.Error("在cses 1确认压缩时Mat转byte失败", thisErr)
			task.ErrMsg = g.UnknownErrorMsg
			return
		}
		defer imgbuf.Close()
		outByte, thisErr := imagequant.Crush(imgbuf.GetBytes(), 10, png.CompressionLevel(compressWay.WayParam))
		if thisErr != nil {
			task.Log.Error("在cses 1使用pngquant压缩时失败", thisErr)
			task.ErrMsg = g.UnknownErrorMsg
			return
		}
		myFileUtil.SaveImg(task, outByte, "water_")
		if task.ErrMsg != "" {
			return
		}
	case 2: // jpg压缩
		if !gocv.IMWriteWithParams(task.ResultPath, task.Mat, []int{gocv.IMWriteJpegQuality, compressWay.WayParam}) {
			task.Log.Error("gocv在comfirmCompress“case 2” 保存失败")
			return
		}
	}
	return nil
}

/**
 * @description: png图像无损压缩
 * @param {*dto.ImgDto} img
 * @param {int} pngCompression 压缩率 范围[0-9],数值越大压缩率越高
 * @return {*} 文件压缩后的大小
 */
func PngLosslessCompress(task *model.ImgTask, pngCompression int) {
	// 这里要先把文件保存到本地，才能处理
	// 加载原始结果图
	var err error
	task.Mat, err = gocv.IMDecode(task.ImgByte, gocv.IMReadUnchanged)
	if err != nil {
		task.Log.Error("解析图片失败：", err)
		task.ErrMsg = "解析图片失败"
		return
	}
	defer task.Mat.Close()
	if task.ResultPath = GocvSaveImg(task, "result_", pngCompression); task.ErrMsg != "" {
		return
	}
}
