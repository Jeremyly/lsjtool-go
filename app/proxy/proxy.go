/*
 * @Author: lsjweiyi
 * @Date: 2024-03-04 18:49:46
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-12 10:28:10
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package proxy

import (
	"path"
	g "src/global"
	"strings"

	"github.com/gin-gonic/gin"
)

/**
 * @description: 静态资源的代理，包含了判断是否支持压缩，是否启用缓存
 * @param {*gin.Context} c
 * @return {*}
 */
func ProxyStatic(c *gin.Context) {

	var file string
	urlPath := c.Request.URL.Path
	// 路径以/结尾，或者直接以一个目录结尾的，默认其请求的是index.html
	if strings.HasSuffix(urlPath, "/") || !strings.Contains(urlPath, ".") {
		file = path.Join(urlPath, "index.html")
		c.Header("Content-Type", `text/html; charset=utf-8`) // 因为后续可能返回它的压缩文件，必须得手工设置类型，否则c.File获取的是压缩文件的类型
	} else { // 其他则是带具体文件名的请求，当前我的文件类型里只有.js、.css、.jpg和.json，如果还有其他，需在此补充
		file = urlPath
		if strings.HasSuffix(urlPath, ".js") {
			c.Header("Content-Type", `application/javascript; charset=utf-8`) // 因为后续可能返回它的压缩文件，必须得手工设置类型，否则c.File获取的是压缩文件的类型
		} else if strings.HasSuffix(urlPath, ".css") {
			c.Header("Content-Type", `text/css; charset=utf-8`) // 因为后续可能返回它的压缩文件，必须得手工设置类型，否则c.File获取的是压缩文件的类型
		} else if strings.HasSuffix(urlPath, ".jpg") {
			c.Header("Content-Type", `image/jpeg`)
		} else if strings.HasSuffix(urlPath, ".png") {
			c.Header("Content-Type", `image/png`)
		} else if strings.HasSuffix(urlPath, ".json") {
			c.Header("Content-Type", `application/json`)
		}
	}

	var filMap map[string]*g.StaticFile
	// 判断是否是手机端访问
	if strings.Contains(c.Request.UserAgent(), "Mobile") {
		filMap = g.H5FlieMap
	} else {
		filMap = g.WebFlieMap
	}

	// 判断文件是否存在，如果不存在则返回404.html
	_, ok := filMap[file]
	if !ok {
		c.Header("Content-Type", `text/html; charset=utf-8`) // 因为后续可能返回它的压缩文件，必须得手工设置类型，否则c.File获取的是压缩文件的类型
		file = `/404.html`
		c.Writer.Write(filMap[file].File)
		return
	} else if c.Request.Header.Get("If-None-Match") == filMap[file].Md5 { // 检查客户端是否启用缓存，如果启用则检查文件是否修改
		writeNotModified(c)
		return
	}
	c.Header("Etag", filMap[file].Md5) // 文件id

	// 下面是判断请求是否支持压缩，如果支持，先判断是否有对应的压缩文件，有则返回压缩文件，无则返回原文件
	acceptEncoding := c.Request.Header.Get("Accept-Encoding")

	if strings.Contains(acceptEncoding, "br") {
		_, ok = filMap[file+".br"]
		if ok {
			c.Header("Content-Encoding", "br")
			file = file + ".br"
		}
	} else if strings.Contains(acceptEncoding, "gzip") {
		_, ok = filMap[file+".gz"]
		if ok {
			c.Header("Content-Encoding", "gzip")
			file = file + ".gz"

		}
	}

	c.Writer.Write(filMap[file].File)
}

/**
 * @description: 文件未修改的响应，删除掉不需要的header
 * @param {*gin.Context} c
 * @return {*}
 */
func writeNotModified(c *gin.Context) {
	delete(c.Writer.Header(), "Content-Type")
	c.Status(304)
}
