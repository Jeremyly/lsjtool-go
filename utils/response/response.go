/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2023-08-26 19:05:57
 * @Description: http请求返回体
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package response

import (
	"net/http"
	g "src/global"

	"github.com/gin-gonic/gin"
)

func ReturnJson(ctx *gin.Context, dataCode int, data interface{}) {
	// 因为 error类型不是结构体，无法转化成json序列，所以需要处理下
	err, ok := data.(error)
	if ok {
		data = err.Error()
	}
	//Context.Header("key2020","value2020")  	//可以根据实际情况在头部添加额外的其他信息
	ctx.JSON(http.StatusOK, gin.H{
		"code": dataCode,
		"data": data,
	})

}

// Success 直接返回成功
func Success(c *gin.Context, data interface{}) {
	ReturnJson(c, g.StatusOkCode, data)
}

// Fail 失败的业务逻辑
func Fail(c *gin.Context, dataCode int, data interface{}) {
	ReturnJson(c, dataCode, data)
	c.Abort()
}
