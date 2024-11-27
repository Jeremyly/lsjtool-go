/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-08 22:20:49
 * @Description:全局变量的定义
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package global

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"gorm.io/gorm"
)

// 这里定义一些全局变量
var (
	BasePath string // 定义项目的根目录

	ZapLog *zap.Logger // 全局日志指针

	VP *viper.Viper // 全局配置文件指针

	Mysql *gorm.DB // 全局gorm的客户端连接

	IP *ipVisitS // ip管理的对象

	BaiduToken = make(map[int]string) // 储存百度token的map

	WechatPay *core.Client // 微信支付客户端

	Watermark1   *gocv.Mat // 保存水印图的索引，避免每次读取文件,1通道的
	WatermarkBW1 *gocv.Mat // 保存水印图的二值图，避免每次读取文件 1通道的
	Watermark3   *gocv.Mat // 保存水印图的索引，避免每次读取文件,3通道的
	WatermarkBW3 *gocv.Mat // 保存水印图的二值图，避免每次读取文件 3通道的
	Watermark4   *gocv.Mat // 保存水印图的索引，避免每次读取文件,4通道的
	WatermarkBW4 *gocv.Mat // 保存水印图的二值图，避免每次读取文件 4通道的

	ZeroStamp int // 今日0点的时间戳

	ExitSignal = make(chan os.Signal, 1)
)

// 1.初始化程序根目录
func init() {

	if curPath, err := os.Getwd(); err == nil {
		// 路径进行处理，兼容单元测试程序程序启动时的奇怪路径
		if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-test") {
			BasePath = strings.Replace(strings.Replace(curPath, `\test`, "", 1), `/test`, "", 1)
		} else {
			BasePath = curPath
		}
	} else {
		log.Fatal("初始化项目根目录失败")
	}
}
