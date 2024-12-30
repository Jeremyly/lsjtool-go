/*
 * @Author: lsjweiyi
 * @Date: 2023-07-23 15:30:59
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-09 21:14:37
 * @Description:该中间件限制整个服务的访问量，利用标准库，性能好，是令牌桶原理
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package middleware

import (
	"fmt"
	g "src/global"
	"src/utils/response"
	"time"

	"github.com/gin-gonic/gin"
)

/**
 * @description: 该中间件控制Ip访问
 * @return {*}
 */
func IPHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := g.IP.Add(c.ClientIP())
		if res == 0 {
			response.Fail(c, g.AccessForbiddenCode, "ip格式错误")
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " ", c.ClientIP(), c.Request.URL.Path, " ")
			c.Abort()
		} else if res == -128 {
			response.Fail(c, g.AccessForbiddenCode, "你已被禁用永久封禁！")
			c.Abort()
		} else if res < 0 {
			response.Fail(c, g.AccessForbiddenCode, fmt.Sprintf("由于访问过于频繁，您已被禁用%d分钟！", -res))
			if res == -127 {
				fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " ", c.ClientIP(), c.Request.URL.Path, " ")
			}
			c.Abort()
		} else {
			c.Next()
			// 如果是300走缓存，则减少访问次数
			if c.Writer.Status() >= 300 && c.Writer.Status() < 400 {
				g.IP.Reduce(c.ClientIP())
			}
		}
	}
}
