package payService

import (
	"errors"
	"src/app/model"
)

// 对工具付款后增加下载次数
func ToolAddDownTimes(order *model.Order) error {
	addDownTimes := 1       // 默认都是1
	if order.ToolId == 27 { // 27号工具每次付款有两次下载机会
		addDownTimes = 2
	}
	if err := model.AddImgTaskDownTimes(order, addDownTimes); err != nil { // 更新任务

		order.Log.Error("付款增加下载次数失败", err)
		return errors.New("付款增加下载次数失败")
	}
	return nil
}

// 对工具退款后减少下载次数
func ToolRefundReduceDownTimes(order *model.Order) error {
	addDownTimes := -1      // 默认都是1
	if order.ToolId == 27 { // 27号工具每次付款有两次下载机会
		addDownTimes = -2
	}
	if err := model.AddImgTaskDownTimes(order, addDownTimes); err != nil { // 更新任务
		order.Log.Error("退款减少下载次数失败", err)
		return errors.New("退款减少下载次数失败")
	}
	return nil
}
