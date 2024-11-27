/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-06 19:27:41
 * @Description: 路由的请求方法封装
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package router

import (
	"src/middleware"

	"github.com/gin-gonic/gin"
)

// 为了减少对g.ValidatorPrefix的引用，这里再封装一层，需要使用参数验证器的请使用这里的请求方法

// 下面两个参数是为了减少重复代码，减少Request的入参
 var validatorRouter *gin.RouterGroup // 记录当前使用的路由
 var Method string                    // 记录当前的请求方法
 
 /**
  * @description: 使用超时中间件的请求。简化绑定路由时需要的参数，仅适用于需要使用参数的形式，如果不需要使用参数校验，则使用基本router的get方法
  * @param {string} url 绑定的路由路径
  * @param {validatorInterface} validate 参数验证器的结构体
  * @param {...gin.HandlerFunc} handler 请求需要绑定的中间件，没有就不填
  * @return {*}
  */
 func RequestTimeout(url string, apiFunc gin.HandlerFunc, handler ...gin.HandlerFunc) {
	newHandler := []gin.HandlerFunc{middleware.TimeoutMiddleware()} // 确保该中间件处于第一个
	newHandler = append(newHandler, handler...) // 类型正确的参数验证器级添加到中间件中，后续就再请求中就会被调用。
	newHandler = append(newHandler, apiFunc) // 类型正确的参数验证器级添加到中间件中，后续就再请求中就会被调用。
 
	 AddRoute(url, newHandler)
 }
 
 /**
  * @description: 没有使用超时中间件的请求
  * @param {string} url
  * @param {validatorInterface} validate
  * @param {...gin.HandlerFunc} handler
  * @return {*}
  */
 func Request(url string, apiFunc gin.HandlerFunc, handler ...gin.HandlerFunc) {
	 handler = append(handler, apiFunc) // 类型正确的参数验证器级添加到中间件中，后续就再请求中就会被调用。
	 AddRoute(url, handler)
 }
 
 /**
  * @description: 请求实现
  * @param {string} url
  * @param {[]gin.HandlerFunc} handler
  * @return {*}
  */
 func AddRoute(url string, handler []gin.HandlerFunc) {
	 switch Method {
	 case "POST":
		 validatorRouter.POST(url, handler...)
	 case "GET":
		 validatorRouter.GET(url, handler...)
	 case "DELETE":
		 validatorRouter.DELETE(url, handler...)
	 case "PUT":
		 validatorRouter.PUT(url, handler...)
	 }
 
	 // 还有其他方法就继续case
 }