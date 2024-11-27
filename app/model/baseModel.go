/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-15 20:02:24
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-12 19:52:01
 * @Description: 数据库的基础模型
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package model

import (
	"context"
	g "src/global"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseModel struct {
	C   *gin.Context    `json:"-" gorm:"-"` // 将 gin 的上下文对象传送引入结构体，方便后续调用
	Log *g.Log          `json:"-" gorm:"-"` // 将 日志对象传送引入结构体，方便记录日志
	Ctx context.Context `json:"-" gorm:"-"` // 将 上下文对象传送引入结构体，方便后续调用
	DB  *gorm.DB        `json:"-" gorm:"-"` // 将 数据库对象传送引入结构体，方便后续调用
}

func InitBaseModel(c *gin.Context) *BaseModel {
	return &BaseModel{C: c}
}

// 获取日志对象，如无需要可以不调用
func (b *BaseModel) GetLog(c *gin.Context) *BaseModel {
	temp, _ := c.Get("log")
	b.Log, _ = temp.(*g.Log)
	return b
}

// 获取上下文对象，如无需要可以不调用
func (b *BaseModel) GetCtx(c *gin.Context, ctx ...context.Context) *BaseModel {
	if len(ctx) > 0 {
		b.Ctx = ctx[0]
		return b
	}
	temp := c.Request.Context()
	b.Ctx = temp
	return b
}

// 调用这个就不需要调用前两个了
func (b *BaseModel) GetDB(c *gin.Context, newCtx ...context.Context) *BaseModel {
	if len(newCtx) > 1 { // 最多输入一个
		return nil
	}
	var ctx context.Context
	if len(newCtx) > 0 {
		ctx = context.WithValue(newCtx[0], g.MysqlLogKeyName, b.Log)
	} else {
		ctx = context.WithValue(c.Request.Context(), g.MysqlLogKeyName, b.Log)
	}
	b.DB = g.Mysql.WithContext(ctx)
	return b
}
