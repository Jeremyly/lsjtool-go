/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-21 19:16:39
 * @Description: 权限校验相关的中间件
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package middleware

import (
	g "src/global"

	"src/utils/response"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

type HeaderParams struct {
	Authorization string `header:"Authorization" binding:"required,min=10"`
}

// CheckCaptchaAuth 验证码中间件
func CheckCaptchaAuth() gin.HandlerFunc {
	captchaIdKey := g.VP.GetString("Captcha.captchaId")
	captchaValueKey := g.VP.GetString("Captcha.captchaValue")
	return func(c *gin.Context) {
		captchaId := c.Param(captchaIdKey)
		value := c.Param(captchaValueKey)
		if captchaId == "" || value == "" {
			response.Fail(c, 400351, "校验验证码：提交的参数无效，请检查 【验证码ID、验证码值】 提交时的键名是否与配置项一致")
			return
		}
		if captcha.VerifyString(captchaId, value) {
			c.Next()
		} else {
			response.Fail(c, 400355, "验证码校验失败")
		}
	}
}
