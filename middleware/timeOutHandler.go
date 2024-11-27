/*
 * @Author: lsjweiyi
 * @Date: 2023-07-29 21:46:28
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-07 21:39:58
 * @Description:该时间超时中间件是对接口进行限制
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
	"golang.org/x/time/rate"
)

/**
 * @description: 超时的返回
 * @param {*gin.Context} c
 * @return {*}
 */
func timeoutResponse(c *gin.Context) {
	c.JSON(http.StatusGatewayTimeout, "当前业务繁忙，请求被强制终止，请稍后再试！")
}

/**
 * @description: 这个中间件是最后的保障，任务异常导致长时间未响应，则直接返回超时信息
 * @return {*}
 */
func TimeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(60*time.Second), //这里设置多少秒取决于长任务的时间，但场任务一般都是开线程后台运行，不会保持http连接执行
		timeout.WithHandler(LeakyBucketHandler()),
		timeout.WithResponse(timeoutResponse),
	)
}

/**
 * @description:这个中间件是为了对某个接口进行限制，防止接口被恶意请求，导致服务器崩溃
 * 限制一个接口一秒的响应次数，当超过限制则会让请求等待，该库采用漏铜策略，即请求的间隔是固定的
 * @return {*}
 */
func LeakyBucketHandler() gin.HandlerFunc {
	var rl = ratelimit.New(1, ratelimit.WithoutSlack) // per second
	return func(c *gin.Context) {
		startTime := time.Now()
		rl.Take() // 如果请求超过限制，会休眠一段时间
		// 设置十秒钟还未响应就拒绝请求了，说明请求等待时间过长，一般用户也不会等待这么久，避免服务器资源浪费，直接抛弃
		if time.Since(startTime) > 10*time.Second {
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}
		c.Next()
	}
}

/**
 * @description: 该中间件为了限制服务整体的访问量，根本目的是怕太多http请求挂着，导致网络阻塞。利用标准库，性能好，是令牌桶原理
 * @return {*}
 */
func LimitHandler() gin.HandlerFunc {
	var lmt = rate.Limit(50)            // 每秒最大访问量
	var lmter = rate.NewLimiter(lmt, 20) // 瞬时最大访问量
	return func(c *gin.Context) {
		if !lmter.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
		} else {
			c.Next()
		}
	}
}
