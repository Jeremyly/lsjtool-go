/*
 * @Author: lsjweiyi
 * @Date: 2023-08-13 14:57:19
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-25 20:45:38
 * @Description: 文件的工具类
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package myFileUtil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"src/app/model"
	g "src/global"
)

// SaveImg 用于保存图片文件
// 参数：
//   - task: 包含图片保存信息的ImgTask指针
//   - fileByte: 图片文件的字节数组
//   - fileNamePrefix: 图片文件名前缀
//
// 返回值：无
func SaveImg(task *model.ImgTask, fileByte []byte, fileNamePrefix string) (saveFile string) {
	savePath := model.ToolsMap[task.ToolId].SaveDir
	dirExist, err := PathExistsAndCreat(savePath)
	if !dirExist {
		task.Log.Error("创建文件保存目录报错：", err)
		task.ErrMsg = "创建文件保存目录报错"
		fileByte = nil
		return
	}
	saveFile = GetImgSaveFullPath(task, fileNamePrefix)
	if err := SaveFileFromByte(saveFile, fileByte); err != nil {
		task.Log.Error("保存文件失败：", err)
		task.ErrMsg = "保存文件失败"
	}
	fileByte = nil // 清空fileByte，释放内存
	return
}

func GetImgSaveFullPath(task *model.ImgTask, fileNamePrefix string) (fullPath string) {
	savePath := model.ToolsMap[task.ToolId].SaveDir
	dirExist, err := PathExistsAndCreat(savePath)
	if !dirExist {
		task.Log.Error("创建文件保存目录报错：", err)
		task.ErrMsg = "创建文件保存目录报错"
		return
	}
	return path.Join(savePath, fmt.Sprintf("%s%s.%s", fileNamePrefix, task.Id, task.Suffix))
}

/**
 * @description: 将文件的二进制保存成文件的封装
 * @param {string} saveFileName 目标文件名
 * @param {[]byte} fileByte 文件的二进制流
 * @return {*}
 */
func SaveFileFromByte(saveFileName string, fileByte []byte) (err error) {
	saveFileOs, err := os.Create(saveFileName)
	if err != nil {
		return err
	}
	defer saveFileOs.Close()
	_, err = saveFileOs.Write(fileByte)
	return err
}

/**
 * @description: 将文件的byte流保存到对应的工具目录下，这里指的是保存源文件
 * @param {context.Context} ctx
 * @param {*log.Log} log
 * @param {[]byte} fileByte
 * @param {int} toolId
 * @param {int} taskId
 * @param {string} suffix 文件的后缀名
 * @return {*}
 */
func SaveFileOnToolFromByte(ctx context.Context, log *g.Log, fileByte []byte, toolId int, taskId int, suffix string) error {
	savePath := model.ToolsMap[uint8(toolId)].SaveDir
	dirExist, err := PathExistsAndCreat(savePath)
	if !dirExist {
		log.Error("创建文件保存目录报错：", err)
		ctx.Done()
		return errors.New("创建文件保存目录报错")
	}
	saveFile := path.Join(savePath, fmt.Sprintf("%d.%s", taskId, suffix))
	return SaveFileFromByte(saveFile, fileByte)
}

/**
 * @description: 根据taskId从工具对应的目录下读取原始文件成byte流
 * @param {context.Context} ctx
 * @param {*log.Log} log
 * @param {int} toolId 工具id
 * @param {int} taskId 任务id
 * @return {*}
 */
func ReadToolRegionFile2Byte(ctx context.Context, log *g.Log, toolId int, taskId int) (fileByte []byte, suffix string, err error) {
	savepath := model.ToolsMap[uint8(toolId)].SaveDir
	err = errors.New(g.UnknownErrorMsg)
	if savepath == "" {
		return
	}
	thisErr := g.Mysql.Table("user").Where("tool_id=? and task_id=?", toolId, taskId).Pluck("suffix", &suffix).Error
	if thisErr != nil {
		log.Error("sql查询后缀名失败", thisErr)
		return
	}
	if suffix == "" {
		log.Error("未查到对应的后缀名")
		err = errors.New("未查到对应的文件，请检查任务id是否正确")
		return
	}
	fileFullName := path.Join(savepath, fmt.Sprintf("%d.%s", taskId, suffix))
	dirExist, thisErr := FileOrPathExists(fileFullName)
	if thisErr != nil {
		log.Error("检索文件报错：", thisErr)
		return
	} else if !dirExist {
		log.Error(fmt.Sprintf("文件：%s 不存在，请检查是否被删除", fileFullName))
		return
	}
	fileByte, thisErr = os.ReadFile(fileFullName)
	if thisErr != nil {
		log.Error("打开文件时报错：", thisErr)
		return
	}
	err = nil
	return
}
