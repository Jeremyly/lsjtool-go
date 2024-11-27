/*
 * @Author: lsjweiyi
 * @Date: 2023-09-14 19:53:36
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-04 18:58:01
 * @Description: 管理定时器
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package timer

import (
	"log"
	"src/global"
	"time"
)

type MyTimer interface {
	Stop() bool
}
type MyTicker interface {
	Stop()
}

var (
	StopChan chan bool = make(chan bool) // 退出定时器所在的协程，用select 配合监听

	TimerList  []MyTimer  = make([]MyTimer, 0)  // 用于保存所有需要在程序结束时关闭的定时器
	TickerList []MyTicker = make([]MyTicker, 0) // 用于保存所有需要在程序结束时关闭的定时器

)

/**
 * @description: 启动定时器
 * @return {*}
 */
func StartTimer() {
	// 初始化百度token
	if err := LoadToken(StopChan); err != nil {
		log.Fatal("加载百度apitoken异常:" + err.Error())
	}
	// 访问量汇总定时器
	ToolUseSummaryTimer()

	// 定时主动查询仍在支付中的订单状态
	QueryOrderPayStatusTicker()

	// 迁移数据定时器
	MigrateDataTicker()

	// 删除保存的文件
	DelSaveFilesTicker()
	
	// 定时检查ip
	ipTicker := global.IP.CheckIPList(StopChan)
	TickerList = append(TickerList, ipTicker)
}

/**
 * @description: 结束定时器
 * @return {*}
 */
func Stop() {
	thisLog := global.Log{RequestUrl: "timerManager.Stop"}
	// 停止定时器
	for _, myTimer := range TimerList {
		myTimer.Stop()
	}
	thisLog.Info("Timer全部关闭", "")
	for _, myTicker := range TickerList {
		myTicker.Stop()
	}
	thisLog.Info("Ticker全部关闭", "")
	// 有多少个定时器就发送多少个关闭信号
	for i := 0; i < len(TickerList)+len(TimerList); i++ {
		StopChan <- true // 定时器stop并不会关闭他们所在的协程，需要额外使用StopChan发出关闭协程的信号
	}
	thisLog.Info("定时器所在的协程全部关闭", "")
	time.Sleep(1 * time.Second) // 延迟一秒退出，给定时器一些时间收尾
}
