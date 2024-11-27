/*
 * @Author: lsjweiyi
 * @Date: 2023-09-13 18:42:05
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 18:15:52
 * @Description: 工具的使用量汇总定时器
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package timer

import (
	g "src/global"
	"src/utils/base"
	"time"
)

/**
 * @description: 汇总工具使用量的定时器
 * @return {*}
 */
func ToolUseSummaryTimer() {
	g.ZeroStamp = base.GetTodayTimeStamp() // 获取当天的0点时间戳
	nextZeroStamp := g.ZeroStamp + 86400   // 加一天的秒数，得到第二天的0点

	hourTicker := time.NewTicker(1 * time.Hour) // 定时器，每小时执行一次
	TickerList = append(TickerList, hourTicker)

	zeroTimer := time.NewTimer(time.Until(time.Unix(int64(nextZeroStamp), 0))) //生成一个明天零点执行的定时器
	TimerList = append(TimerList, zeroTimer)                                   // 添加到队列，方便结束程序时退出定时器
	go func() {
		for {
			select {
			case <-zeroTimer.C:
				base.Summary(g.ZeroStamp)              //同步数据
				g.ZeroStamp = base.GetTodayTimeStamp() // 进入里面，说明到第二天了，更新为当天的0点
				nextZeroStamp += 86400
				zeroTimer.Reset(time.Until(time.Unix(int64(nextZeroStamp), 0)))
			case <-hourTicker.C: // 每小时汇总一次数据
				base.Summary(g.ZeroStamp) // 汇总数据
			case <-StopChan:
				// 程序退出前，将工具使用量汇总
				base.Summary(g.ZeroStamp) //同步数据
				return                    // 退出协程
			}
		}
	}()
	// 下面这个协程是为了对照hourTicker，因为程序结束时，是按照定时器的数量发送StopChan的，而hourTicker和zeroTimer作用于同一个协程，导致少了一个协程接收StopChan信号，所以增加一个协程来接收
	go func() {
		<-StopChan
	}()
}
