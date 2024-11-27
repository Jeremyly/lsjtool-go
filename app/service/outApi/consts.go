/*
 * @Author: lsjweiyi
 * @Date: 2023-07-09 13:57:49
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 12:22:50
 * @Description: 百度api相关常量
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package outApi

import (
	"src/global"
)

var (
	// 人像分割
	PortraitAppId  int
	PortraitApiUrl string
	// 图像特效与增强
	EffectsEnhancementAppId  int
	EffectsEnhancementApiUrl string
	// 文档图像处理
	DocumentAppId  int
	DocumentApiUrl string
)

// 初始化
func InitBaiduApp() {
	// 人像分割
	PortraitAppId = global.VP.GetInt("BaiduApp.PortraitAppId")
	PortraitApiUrl = global.VP.GetString("BaiduApp.PortraitApiUrl")
	// 图像特效与增强
	EffectsEnhancementAppId = global.VP.GetInt("BaiduApp.EffectsEnhancementAppId")
	EffectsEnhancementApiUrl = global.VP.GetString("BaiduApp.EffectsEnhancementApiUrl")
	// 文档图像处理
	DocumentAppId = global.VP.GetInt("BaiduApp.DocumentAppId")
	DocumentApiUrl = global.VP.GetString("BaiduApp.DocumentApiUrl")

	if PortraitAppId == 0 || EffectsEnhancementAppId == 0 || DocumentAppId == 0 || PortraitApiUrl == "" || EffectsEnhancementApiUrl == "" || DocumentApiUrl == "" {
		panic("请检查配置文件BaiduApp相关参数")
	}

}
