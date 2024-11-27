/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-15 22:39:45
 * @Description: 各接口的路由实现
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package router

import (
	"src/app/api"
	"src/middleware"

	"github.com/gin-gonic/gin"
)

// 路由全往这里写
func addTimeoutRoutes(router *gin.RouterGroup) {
	// 下面的路由是有顺序要求的，要严格按照顺序添加

	router.Use(middleware.TimeoutMiddleware(), middleware.LimitHandler(), middleware.LogHandler())

	// 以下路由不需要有task-id
	validatorRouter = router.Group("tools")
	{
		Method = "POST"
		Request("/order/wechatNotify", api.WechatpayNotify)             // 微信支付回调接口
		Request("/order/wechatRefundNotify", api.WechatpayRefundNotify) // 微信支付退款回调接口
		Request("order/aliPayNotify", api.AlipayNotify)                 // 支付宝支付回调接口
		Method = "GET"
		Request("/tool", api.GetToolMsg)                 // 查询工具信息
		Request("/order/status/:orderId", api.PayStatus) // 查询订单支付状态
		Request("/imgDown/:toolId/:taskId", api.ImageDown) // 普通下载接口
	}

	// 以下路由是必须输入taskID的
	validatorRouter = router.Group("tools", middleware.CheckTaskId())
	{
		// get方法写在这下面，以此类推
		Method = "GET"
		// Request("/img2docDownType", api.Get27ToolDownType)  // 图片转文档可供下载的类型
		Request("/img2docDown/:type", api.Get27ToolDownUrl) // 图片转文档可供下载的类型
		Request("/img2docResult", api.DocConvertResult)     // 查询图片转文档结果接口
		Request("/order/wechatOrder", api.WechatPay)        // 创建订单接口
		Request("/order/AliH5Order", api.AliH5Pay)              // 创建订单接口
		Request("/downTimes", api.GetDownTimes)             // 查询下载次数接口
		Method = "POST"                                     // post方法写在这下面
		Request("/img2doc", api.DocImgConvert)              // 图片转文档接口
	}
	//以下任务是可以不输入taskID的
	validatorRouter = router.Group("tools", middleware.CreateTask())
	{
		Method = "POST" // post方法写在这下面
		Request("/cardChangeBGColor", api.CardChangeBGColor)
		Request("/imgLossyCompress", api.LossyCompression)
		Request("/pngLosslessCompress", api.PngLosslessCompress)
		Request("/imgEffectEnhance", api.EffectsEnhancement) // 图像特效处理接口
		Request("/imgConvers", api.ImgConvers)               // 图像转换接口
		// get方法写在这下面，以此类推
		Method = "GET"
		Request("/order/wechatOrderFirst", api.WechatPayFirst) // 创建订单接口，应用与先付款再使用的场景
		Request("/order/AliOrderFirst", api.AliPayFirst)       // 创建订单接口，应用与先付款再使用的场景
	}

}
