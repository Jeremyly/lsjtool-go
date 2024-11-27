// 用于定时查询还在支付中状态的订单定时器，因为回调可能会失败，需要主动查询
package timer

import (
	"src/app/api"
	"src/app/model"
	"src/app/service/payService"
	g "src/global"
	"time"
)

func QueryOrderPayStatusTicker() {
	fiveMinTicker := time.NewTicker(5 * time.Minute) // 定时器，每五分钟执行一次
	TickerList = append(TickerList, fiveMinTicker)   // 将定时器添加到列表中
	go func() {
		for {
			select {
			case <-fiveMinTicker.C: // 定时器触发
				QueryOrderPayStatus()        // 执行订单查询
				api.QueryOrderAutoRefund27() // 执行27工具自动退款
			case <-StopChan:
				return // 退出协程
			}
		}
	}()
}

func QueryOrderPayStatus() {
	thisLog := g.Log{RequestUrl: "QueryOrderPayStatus"}
	// 先从数据库中查询支付状态为1且订单创建时间超过五分钟的订单，然后主动发起订单查询
	// 查询结果根据订单状态进行处理，比如更新数据库中的订单状态
	var orders []model.Order
	err := g.Mysql.Table("order").Where("result_type = ? and created_at < ?", 1, time.Now().Add(-5*time.Minute).UnixMilli()).Find(&orders).Error
	if err != nil {
		thisLog.Error("查询订单失败", err)
		return
	}
	// 发起订单查询
	for _, order := range orders {
		order.DB = g.Mysql // 这里需要手工初始化
		order.Log = &thisLog
		if order.Platform == 1 {
			payService.WechatQueryOrderStatus(order.Id)
		} else if order.Platform == 2 {
			payService.AliQueryOrderStatus(&order)
		}
	}
}
