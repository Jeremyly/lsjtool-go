/*
 * @Author: lsjweiyi
 * @Date: 2023-09-12 18:51:04
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 14:19:33
 * @Description: 记录工具使用量的实现
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package base

import (
	g "src/global"
	"time"
)

// 同步tools_data_use
type toolsDataUseS struct {
	ToolId       uint8
	Date         int
	Call         int
	Success      int
	Pay          int
	PayAmount    int64
	Refund       int
	RefundAmount int64
	Download     int
	Api          int
}

var (
	toolUseSumLock = make(chan bool, 1) // 锁,防止并发操作
	toolUseSum     = make(map[int]map[uint8]*toolsDataUseS)
)

/**
 * @description: 获取当天0点的时间戳
 * @return {*}
 */
func GetTodayTimeStamp() (timestamp int) {
	// 获取当前时间
	now := time.Now()
	// 设置当天00:00:00的时间
	zero := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	// 计算今天整点的时间戳
	timestamp = int(zero.Unix())
	return
}

/**
 * @description: 工具调用一次加一
 * @param {int} toolId
 * @return {*}
 */
func ToolCall(toolId uint8) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.Call++
	<-toolUseSumLock
}

/**
 * @description: 工具调用成功一次加一
 * @param {int} toolId
 * @return {*}
 */
func ToolCallSuccess(toolId uint8) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.Success++
	<-toolUseSumLock
}

/**
 * @description: 工具成功付费一次加一
 * @param {int} toolId
 * @return {*}
 */
func ToolPay(toolId uint8) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.Pay++
	<-toolUseSumLock
}

/**
 * @description: 工具付费金额累加
 * @param {int} toolId
 * @return {*}
 */
func ToolPayAmount(toolId uint8, amount int64) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.PayAmount += amount
	<-toolUseSumLock
}

/**
 * @description: 工具退款一次加一
 * @param {int} toolId
 * @return {*}
 */
func ToolRefund(toolId uint8) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.Refund++
	<-toolUseSumLock
}

/**
 * @description: 工具付费金额累加
 * @param {int} toolId
 * @return {*}
 */
func ToolRefundAmount(toolId uint8, amount int64) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.RefundAmount += amount
	<-toolUseSumLock
}

/**
* @description: 工具结果下载一次加一
* @param {int} toolId
* @return {*}
 */
func ToolDownload(toolId uint8) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.Download++
	<-toolUseSumLock
}

/**
 * @description: 工具依赖的外部api成功调用一次加一
 * @param {int} toolId
 * @return {*}
 */
func OutApiCall(toolId uint8) {
	toolUseSumLock <- true
	a := findOrCreate(toolId)
	a.Api++
	<-toolUseSumLock
}

func findOrCreate(toolId uint8) *toolsDataUseS {
	if toolUseSum[g.ZeroStamp] == nil {
		toolUseSum[g.ZeroStamp] = make(map[uint8]*toolsDataUseS)
	}
	if toolUseSum[g.ZeroStamp][toolId] == nil {
		toolUseSum[g.ZeroStamp][toolId] = &toolsDataUseS{
			ToolId: toolId,
			Date:   g.ZeroStamp,
		}
	}
	return toolUseSum[g.ZeroStamp][toolId]
}

/**
 * @description: 汇总工具使用量，记录到mysql中
 * @return {*}
 */
func Summary(timeStap int) {
	toolUseSumLock <- true
	for _, v := range toolUseSum[timeStap] {
		var thisToolDataUse toolsDataUseS
		// 查询是否有相同的数据已经存在数据库了，依据双主键查询的，如果有，也仅有一条
		result := g.Mysql.Table("tools_data_use").Where("tool_id=? and date=?", v.ToolId, v.Date).Take(&thisToolDataUse)
		if result.RowsAffected == 0 { // 没查到就直接插入现有数据
			g.Mysql.Table("tools_data_use").Create(v)
		} else { //查到了则将查到的值加上现在的值，一起更新，说明当天已更新过
			v.Call += thisToolDataUse.Call
			v.Success += thisToolDataUse.Success
			v.Pay += thisToolDataUse.Pay
			v.PayAmount += thisToolDataUse.PayAmount
			v.Refund += thisToolDataUse.Refund
			v.RefundAmount += thisToolDataUse.RefundAmount
			v.Download += thisToolDataUse.Download
			v.Api += thisToolDataUse.Api
			g.Mysql.Table("tools_data_use").Where("tool_id=? and date=?", v.ToolId, v.Date).Updates(v)
		}
	}
	delete(toolUseSum, timeStap)
	<-toolUseSumLock
}
