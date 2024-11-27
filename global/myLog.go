/*
 * @Author: lsjweiyi
 * @Date: 2023-09-12 18:31:17
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-21 22:00:03
 * @Description: 封装自己的日志结构体
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package global

import (
	"go.uber.org/zap"
)

// 封装一个结构体
type Log struct {
	Logid      *int
	RequestUrl string
}

func InitLog(logId int) *Log {
	return &Log{Logid: &logId}
}

// 以下方法将logid封装进去，使用时可以做到无感知
func (l *Log) Info(msg string, info interface{}) {
	ZapLog.Info(msg, zap.Intp(LogIdKey, l.Logid), zap.String("RequestUrl", l.RequestUrl), zap.Any("info", info))
}
func (l *Log) Warn(msg string, warn interface{}) {
	ZapLog.Warn(msg, zap.Intp(LogIdKey, l.Logid), zap.String("RequestUrl", l.RequestUrl), zap.Any("warn", warn))
}
func (l *Log) Error(msg string, err ...error) {
	ZapLog.Error(msg, zap.Intp(LogIdKey, l.Logid), zap.String("RequestUrl", l.RequestUrl), zap.Errors("error", err))
}
