/*
 * @Author: lsjweiyi
 * @Date: 2023-06-29 20:55:52
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-03 15:43:09
 * @Description:图像相关的，在service这里用得着的工具类
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */

package service

import (
	"encoding/base64"
	"fmt"
	"image"
	"math"
	"src/app/model"
	"src/global"
	g "src/global"
	myFileUtil "src/utils/file"
	"sync"

	"github.com/h2non/filetype"
	"gocv.io/x/gocv"
)

/**
 * @description: base64字符串转Rgb模式gocv.mat
 * @param {string} imgStr 图片的base64
 * @return {*}
 */
func ImgStr2RGB(imgStr string) (gocv.Mat, error) {
	imgMat := gocv.NewMat()
	// base64解码
	img, err := base64.StdEncoding.DecodeString(imgStr)
	if err != nil {
		return imgMat, err
	}
	imgMat, err = gocv.IMDecode(img, gocv.IMReadUnchanged)
	if err != nil {
		return imgMat, err
	}
	return imgMat, nil
}

/**
 * @description: 给图片添加固定的水印，这样的性能很高，通常不超过3ms
 * @param {*gocv.Mat} img
 * @return {*}
 */
func ImgWatermark(img gocv.Mat) {
	var Watermark gocv.Mat
	var WatermarkBW gocv.Mat
	// 不同通道的图片使用不同的水印
	if img.Channels() == 3 {
		// 判断是否需要初始化
		if g.Watermark3 == nil {
			mat := gocv.IMRead(global.VP.GetString("WaterImg"), gocv.IMReadColor)
			g.Watermark3 = &mat

			// 取得水印图片的二值图
			watermarkGray3 := gocv.NewMat()
			gocv.CvtColor(*g.Watermark3, &watermarkGray3, gocv.ColorBGRToGray)
			// 取得二值图的黑白图
			BW3 := gocv.NewMat()
			g.WatermarkBW3 = &BW3
			gocv.Threshold(watermarkGray3, g.WatermarkBW3, 30, 255, gocv.ThresholdBinary)
			watermarkGray3.Close()
		}
		Watermark = *g.Watermark3
		WatermarkBW = *g.WatermarkBW3
	} else if img.Channels() == 4 {
		// 判断是否需要初始化
		if g.Watermark4 == nil {
			mat := gocv.IMRead(global.VP.GetString("WaterImg"), gocv.IMReadColor)
			g.Watermark4 = &mat
			gocv.Merge([]gocv.Mat{mat, gocv.NewMatWithSizeFromScalar(gocv.NewScalar(255, 255, 255, 255), mat.Rows(), mat.Cols(), gocv.MatTypeCV8UC1)}, g.Watermark4)
			// 判断是否需要初始化
			findOrCreateWatermark1()
			// 通过单通道二值图拼接出四通道的二值图
			BW4 := gocv.NewMat()
			g.WatermarkBW4 = &BW4
			gocv.Merge([]gocv.Mat{*g.WatermarkBW1, *g.WatermarkBW1, *g.WatermarkBW1, *g.WatermarkBW1}, g.WatermarkBW4)
		}
		Watermark = *g.Watermark4
		WatermarkBW = *g.WatermarkBW4
	} else if img.Channels() == 1 {
		// 判断是否需要初始化
		findOrCreateWatermark1()
		Watermark = *g.Watermark1
		WatermarkBW = *g.WatermarkBW1
	} else {
		fmt.Println("图片通道数不支持", img.Channels())
		return
	}
	waterCol := Watermark.Cols()
	waterRow := Watermark.Rows()
	// 为了水印不要那么密，这里设置一个间距
	spacing := 200 // 水印的间距

	waterColAndSpacing := waterCol + spacing
	watreRowAndSpacing := waterRow + spacing

	// 计算需要添加水印的次数
	watermarkHugeRowTimes := int(math.Ceil(float64(img.Rows()) / float64(watreRowAndSpacing))) //向上取整
	watermarkHugeColTimes := int(math.Ceil(float64(img.Cols()) / float64(waterColAndSpacing)))

	imgRect := image.Rectangle{} //每次取图像中哪一块区域的数据结构
	var croppedMat gocv.Mat      // 每次从原图像上裁剪和水印图像一样大小的一块

	// 循环一块块的给图像添加水印
	for i := 0; i < watermarkHugeColTimes; i++ {
		newWater := Watermark // 因为有可能要裁剪后才可以进行复制，所以不能直接在原图上裁剪，得复制一个新的，下同
		newWaterBW := WatermarkBW
		imgRect.Min.X = i * waterColAndSpacing
		imgRect.Max.X = imgRect.Min.X + waterCol
		if imgRect.Max.X >= img.Cols() { // 判断是否超出原图像的列数
			imgRect.Max.X = img.Cols()
		}
		for j := 0; j < watermarkHugeRowTimes; j++ {
			imgRect.Min.Y = j * watreRowAndSpacing
			imgRect.Max.Y = imgRect.Min.Y + waterRow
			if imgRect.Max.Y >= img.Rows() { // 判断是否超出原图像的行数
				imgRect.Max.Y = img.Rows()
			}
			croppedMat = img.Region(imgRect) // 原图像中一块区域的浅拷贝，修改它会连带修改原图像

			// 判断裁剪的图像大小是否与水印图像大小一致，不一致则需要重新裁剪
			if croppedMat.Rows() != Watermark.Rows() || croppedMat.Cols() != Watermark.Cols() {
				newRect := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: croppedMat.Cols(), Y: croppedMat.Rows()}}
				newWater = Watermark.Region(newRect)
				newWaterBW = WatermarkBW.Region(newRect)
			}

			newWater.CopyToWithMask(&croppedMat, newWaterBW) // 用水印图覆盖原图像
			croppedMat.Close()
		}
	}
}

// 检查单通道的水印图是否生成了
func findOrCreateWatermark1() {
	// 判断是否需要初始化
	if g.Watermark1 == nil {
		mat := gocv.IMRead(global.VP.GetString("WaterImg"), gocv.IMReadGrayScale) // 直接读取黑白
		g.Watermark1 = &mat
		// 取得水印图片的二值图
		BW1 := gocv.NewMat()
		g.WatermarkBW1 = &BW1
		gocv.Threshold(mat, g.WatermarkBW1, 30, 255, gocv.ThresholdBinary)
	}
}

/**
 * @description: 保存图片
 * @param {*model.ImgTask} task
 * @param {string} fileName
 * @param {*gocv.Mat} mat
 * @param {int} quality 保存图像的质量，JPG 的范围是[0-100],png的范围是[0-9],WEBP的范围是[0-100]
 * @return {*}
 */
func GocvSaveImg(task *model.ImgTask, fileNamePrefix string, quality int) (fullPath string) {
	fullPath = myFileUtil.GetImgSaveFullPath(task, fileNamePrefix)
	if task.ErrMsg != "" {
		return
	}
	var saveResult bool
	if quality != -1 {
		switch task.Suffix {
		case "jpg":
			// 保存图片
			saveResult = gocv.IMWriteWithParams(fullPath, task.Mat, []int{gocv.IMWriteJpegQuality, quality})
		case "png":
			// 保存图片
			saveResult = gocv.IMWriteWithParams(fullPath, task.Mat, []int{gocv.IMWritePngCompression, quality})
		case "webp":
			saveResult = gocv.IMWriteWithParams(fullPath, task.Mat, []int{gocv.IMWriteWebpQuality, quality})
		default:
			// 保存图片
			saveResult = gocv.IMWrite(fullPath, task.Mat)
		}
	} else { // -1 表示使用默认方法保存
		// 默认方法保存图片
		saveResult = gocv.IMWrite(fullPath, task.Mat)
	}
	if !saveResult {
		task.ErrMsg = g.UnknownErrorMsg
		return fullPath
	}
	return
}

/**
 * @description: 检查用户上传的图片是否正确
 * @param {model.Task} task 本次任务的对象，内涵多个有用结构体的引用
 * @param {int64} maxSize 本工具允许上传的图像大小
 * @return {imgByte}  图像的二进制
 * @return {suffix} 解析出来的图像格式
 */
func CheckImage(task *model.ImgTask, maxSize int64) (imgByte []byte, suffix string) {
	if task.C.Request.ContentLength > maxSize {
		task.ErrMsg = fmt.Sprintf("文件大小超过限制，最大不得超过 %d KB", maxSize/1024)
		return
	}
	var thisErr error
	// 检测是否正确的上传文件
	if thisErr = task.C.Request.ParseMultipartForm(task.C.Request.ContentLength); thisErr != nil {
		task.Log.Error("未正确上传文件：", thisErr)
		task.ErrMsg = "未正确上传文件"
		return
	}
	// 表明未上传文件
	if len(task.C.Request.MultipartForm.File["file"]) <= 0 {
		return
	}
	// 读取文件
	fileRead, file, thisErr := task.C.Request.FormFile("file")
	if thisErr != nil {
		task.Log.Error("上传的文件错误或文件未上传：", thisErr)
		task.ErrMsg = "上传的文件错误或文件未上传"
		return
	}

	defer fileRead.Close()
	imgByte = make([]byte, file.Size)
	n, thisErr := fileRead.Read(imgByte)
	if thisErr != nil { // 图像不会小于5个byte
		task.ErrMsg = "上传的文件错误！"
		return
	}
	if n < 5 { // 图像不会小于5个byte
		task.ErrMsg = "上传的文件错误！"
		return
	}
	//检查文件格式是不是图像
	thisFileType, thisErr := filetype.Image(imgByte)
	if thisErr != nil {
		task.Log.Error("上传的文件类型识别报错：", thisErr)
		task.ErrMsg = "上传的文件类型识别错误！"
		return
	}
	suffix = thisFileType.Extension
	return
}

var (
	downLock = sync.Mutex{}
)

/**
 * @description: 图片下载的实现
 * @param {*model.ImgTask} task
 * @return {*}
 */
func DownImg(task *model.ImgTask, isReduceDownTimes bool) {
	downLock.Lock()
	defer downLock.Unlock()
	task.FindWithCreateTime(86400)
	if task.ErrMsg != "" {
		return
	}
	if task.DownloadTimes <= 0 {
		task.ErrMsg = "下载次数已用完"
		return
	}
	if isReduceDownTimes {
		if task.ReduceDownloadTimes(); task.ErrMsg != "" {
			return
		}
	}
}
