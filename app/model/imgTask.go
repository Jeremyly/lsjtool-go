/*
 * @Author: lsjweiyi
 * @Date: 2023-07-16 17:23:59
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-03 16:38:55
 * @Description: user 表
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package model

import (
	"context"
	"database/sql"
	"errors"
	g "src/global"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImgTask struct {
	BaseModel
	Id            string `gorm:"primaryKey"`
	ToolId        uint8  `gorm:"primaryKey"`
	Suffix        string // 如果结果是后缀，怎保存文件的后缀名，否则为空
	DownloadTimes int
	Result        string   // 记录一些不同的结果信息
	CreatedAt     int64    `gorm:"autoUpdateTime:milli"`
	UpdatedAt     int64    `gorm:"autoUpdateTime:milli"`
	ImgByte       []byte   `json:"-" gorm:"-"` // 通常保存用户上传的图像
	Mat           gocv.Mat `json:"-" gorm:"-"` // 通常byte转化成OpenCV mat的对象
	ReginSuffix   string   `json:"-" gorm:"-"` // 保存用户上传的图像的格式后缀
	ErrMsg        string   `json:"-" gorm:"-"` // 过程中的错误信息
	ResultPath    string   `json:"-" gorm:"-"` // 保存结果的路径
}

// 初始化
func InitImgTask(c *gin.Context, ctx ...context.Context) *ImgTask {
	toolId, err := strconv.Atoi(c.Request.Header.Get(g.ToolIdKey))
	if err != nil {
		return &ImgTask{ErrMsg: "工具id输入错误"}
	}
	_, ok := ToolsMap[uint8(toolId)]
	if !ok {
		return &ImgTask{ErrMsg: "工具不存在"}
	}
	id := c.Request.Header.Get(g.TaskIdKey)
	return &ImgTask{Id: id, ToolId: uint8(toolId), BaseModel: *InitBaseModel(c).GetCtx(c, ctx...).GetLog(c).GetDB(c, ctx...)}
}

func (t *ImgTask) Add() {
	if err := t.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(t).Error; err != nil {
		t.Log.Error("插入任务到数据库失败：", err)
		t.ErrMsg = "任务出现不可预知的错误导致失败，请将错误截图发给系统管理员"
		return
	}
}

func (t *ImgTask) Find() {
	if t.Id == "" {
		t.ErrMsg = "资源ID不能为空！"
		return
	}
	if err := t.DB.Find(&t).Error; err != nil {
		t.Log.Error("查找资源失败：", err)
		t.ErrMsg = g.UnknownErrorMsg
		return
	}
	if t.CreatedAt == 0 {
		t.ErrMsg = "未找到该资源！"
		return
	}
}

// 更新task
func (t *ImgTask) Update() {
	if err := t.DB.Save(t).Error; err != nil {
		t.Log.Error("更新任务失败：", err)
		t.ErrMsg = g.UnknownErrorMsg
		return
	}
}

// 只能查询创建时间在N秒之内的
func (t *ImgTask) FindWithCreateTime(second int64) {
	if t.Id == "" {
		t.ErrMsg = "资源ID不能为空！"
		return
	}
	if err := t.DB.Where("created_at > ?", (time.Now().Add(-time.Second * time.Duration(second)).UnixMilli())).Find(&t).Error; err != nil {
		t.Log.Error("查找资源失败：", err)
		t.ErrMsg = g.UnknownErrorMsg
		return
	}
	if t.CreatedAt == 0 {
		t.ErrMsg = "未找到该资源！"
		return
	}
}

/**
 * @description: 减少下载次数
 * @return {*}
 */
func (t *ImgTask) ReduceDownloadTimes() {
	// 开启事务，防止并发导致数据异常
	spts := sql.TxOptions{Isolation: sql.LevelSerializable}
	thisMysql := t.DB.Begin(&spts)
	if err := thisMysql.Model(&t).UpdateColumn("download_times", gorm.Expr("download_times - ?", 1)).Error; err != nil {
		t.ErrMsg = g.UnknownErrorMsg
		thisMysql.Rollback()
		return
	}
	thisMysql.Commit()
	t.DownloadTimes -= 1
}

// // 增加下载次数
func AddImgTaskDownTimes(order *Order, addDownTimes int) error {
	if err := order.DB.Table("img_task").Where("id = ?", order.TaskId).UpdateColumn("download_times", gorm.Expr("download_times + ?", addDownTimes)).Error; err != nil {
		order.Log.Error(" 增加下载次数失败：", err)
		return errors.New(" 增加下载次数失败")
	}
	return nil
}
