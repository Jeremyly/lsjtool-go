/*
 * @Author: lsjweiyi
 * @Date: 2024-04-29 22:03:47
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-08 22:39:37
 * @Description: 一些基础的中间件
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package middleware

import (
	"fmt"
	"net/http"
	g "src/global"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"github.com/yitter/idgenerator-go/idgen"
)

var secureMiddle = secure.New(secure.Options{
	SSLRedirect: true, //只允许https请求
	// SSLHost:              "toolsj.cn", //http到https的重定向
	STSSeconds:           15768000, //Strict-Transport-Security header的时效:1年
	STSIncludeSubdomains: true,     //includeSubdomains will be appended to the Strict-Transport-Security header
	STSPreload:           true,     //STS Preload(预加载)
	FrameDeny:            true,     //X-Frame-Options 有三个值:DENY（表示该页面不允许在 frame 中展示，即便是在相同域名的页面中嵌套也不允许）、SAMEORIGIN、ALLOW-FROM uri
	ContentTypeNosniff:   true,     //禁用浏览器的类型猜测行为,防止基于 MIME 类型混淆的攻击
	BrowserXssFilter:     true,     //启用XSS保护,并在检查到XSS攻击时，停止渲染页面
	IsDevelopment:        false,    //开发模式
})

/**
 * @description: 解决跨域问题
 * @return {*}
 */
func Cors() gin.HandlerFunc {
	allowHeaders := fmt.Sprintf("Access-Control-Allow-Headers,%s,%s", g.TaskIdKey, g.ToolIdKey)
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", c.GetHeader("origin"))
		c.Header("Access-Control-Allow-Headers", allowHeaders)
		c.Header("Access-Control-Allow-Credentials", "true")
		// 放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusAccepted)
			return
		}
		c.Next()
	}
}

/**
 * @description: 启用https
 * @return {*}
 */
func HttpsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 如果不安全，终止.
		if err := secureMiddle.Process(ctx.Writer, ctx.Request); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, "数据不安全")
			return
		}
		// 如果是重定向，终止
		if status := ctx.Writer.Status(); status > 300 && status < 399 {
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

/**
 * @description: 为一次请求生成唯一的日志id，方便微服务的日志追踪 中间件
 * @return {*}
 */
func LogHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := int(idgen.NextId())
		c.Writer.Header().Add("Access-Control-Expose-Headers", g.LogIdKey)
		c.Set("log", &g.Log{Logid: &requestId,RequestUrl: c.Request.URL.String()})                   // 将日志方法封装到请求头中
		c.Request.Header.Set(g.LogIdKey, strconv.Itoa(requestId)) // 设置request的header
		c.Header(g.LogIdKey, strconv.Itoa(requestId))             // 设置response的header
		c.Next()
	}
}

// 只允许 127.0.0.1通过的中间件
func LocalhostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ClientIP() != "127.0.0.1" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Unauthorized")
			return
		}
		c.Next()
	}
}
