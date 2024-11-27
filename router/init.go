/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-08 22:34:44
 * @Description: 路由初始化
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package router

import (
	g "src/global"
	"src/middleware"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	var router *gin.Engine
	// 非调试模式（生产模式） 日志写到日志文件
	if !g.VP.GetBool("AppDebug") {

		//【生产模式】
		// 根据 gin 官方的说明：[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
		// 如果部署到生产环境，请使用以下模式：
		// 1.生产模式(release) 和开发模式的变化主要是禁用 gin 记录接口访问日志，
		// 2.go服务就必须使用nginx作为前置代理服务，这样也方便实现负载均衡
		// 3.如果程序发生 panic 等异常使用自定义的 panic 恢复中间件拦截、记录到日志
		router = ReleaseRouter()
		// pprof.Register(router) // 临时测试
	} else {
		router = gin.Default()
		// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
		pprof.Register(router)
	}

	// 设置可信任的代理服务器列表,gin (2021-11-24发布的v1.7.7版本之后出的新功能)
	_ = router.SetTrustedProxies(nil)

	addLocalR(router.Group("")) // 本地开发路由

	router.Use(middleware.IPHandler()) //在此加入的中间件，无论是否有对应路由都会进入
	if !g.VP.GetBool("AppDebug") {
		router.Use(middleware.HttpsHandler()) // 生产模式启用https
	}
	addStaticRoutes(router.Group("")) // 静态资源路由

	// 因为options请求不会进入以下路由内部，所以内部use Cors()是无效的，必须得在最外层应用，直接作用于router的中间件是全局的，无论是否有对应的路由，他都会进入
	router.Use(middleware.Cors())
	addTimeoutRoutes(router.Group(""))
	return router
}
