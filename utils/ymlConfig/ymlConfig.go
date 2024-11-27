/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-21 21:51:40
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package ymlConfig

import (
	g "src/global"

	"github.com/spf13/viper"
)

// 由于 vipver 包本身对于文件的变化事件有一个bug，相关事件会被回调两次
// 常年未彻底解决，相关的 issue 清单：https://github.com/spf13/viper/issues?q=OnConfigChange
// 设置一个内部全局变量，记录配置文件变化时的时间点，如果两次回调事件事件差小于1秒，我们认为是第二次回调事件，而不是人工修改配置文件
// 这样就避免了 viper 包的这个bug

// CreateYamlFactory 创建一个yaml配置文件工厂
// 参数设置为可变参数的文件名，这样参数就可以不需要传递，如果传递了多个，我们只取第一个参数作为配置文件名
func CreateYamlFactory() (*viper.Viper,error) {

	yamlConfig := viper.New()
	// 配置文件所在目录
	yamlConfig.AddConfigPath(g.BasePath + "/config")
	// 需要读取的文件名,默认为：config
	yamlConfig.SetConfigName("config")
	//设置配置文件类型(后缀)为 yml
	yamlConfig.SetConfigType("yml")

	if err := yamlConfig.ReadInConfig(); err != nil {
		return nil,err
	}
	return yamlConfig,nil
}