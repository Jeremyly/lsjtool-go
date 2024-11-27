/*
 * @Author: lsjweiyi
 * @Date: 2024-05-02 14:56:21
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-07 20:31:47
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package global

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type StaticFile struct {
	File []byte
	Md5  string
}

var WebFlieMap = make(map[string]*StaticFile) // 加载前端静态文件到内存，以便快速读取
var H5FlieMap = make(map[string]*StaticFile)  // 加载前端静态文件到内存，以便快速读取

func LoadWebStatic() error {
	if err := LoadWeb(WebFlieMap, VP.GetString("WebPackageName.Web")); err != nil {
		return err
	}
	if err := LoadWeb(H5FlieMap, VP.GetString("WebPackageName.H5")); err != nil {
		return err
	}
	return nil
}

/**
 * @description: 加载静态资源到内存
 * @param {map[string]*StaticFile} fileMap
 * @return {*}
 */
func LoadWeb(fileMap map[string]*StaticFile, pathName string) error {
	// 遍历目录
	err := filepath.Walk("./"+pathName+"/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略目录本身，只处理文件
		if !info.IsDir() {
			// 移除根目录路径
			relativePath := strings.ReplaceAll(strings.TrimPrefix(path, pathName), "\\", "/")
			// 读取文件内容
			var thisFile StaticFile
			thisFile.File, err = os.ReadFile(path)
			sum := md5.Sum(thisFile.File)
			thisFile.Md5 = hex.EncodeToString(sum[:])
			if err != nil {
				return err
			}
			// 将文件路径和内容存入map
			fileMap[relativePath] = &thisFile

		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

/**
 * @description: 更新静态文件的接口
 * @param {*gin.Context} c
 * @return {*}
 */
func UpdateStaticFile(c *gin.Context) {
	WebFlieMapNew := make(map[string]*StaticFile) // 先用一个新的map，防止并发问题
	if err := LoadWeb(WebFlieMapNew, VP.GetString("WebPackageName.Web")); err != nil {
		thislog := Log{RequestUrl: "UpdateStaticFile/LoadWeb/web"}
		thislog.Error("更新静态文件失败", err)
		c.Data(200, "text/html", []byte(err.Error()))
		c.Abort()
		return
	}
	WebFlieMap = WebFlieMapNew // 然后直接覆盖原来的map
	// 更新H5
	H5FlieMapNew := make(map[string]*StaticFile)
	if err := LoadWeb(H5FlieMapNew, VP.GetString("WebPackageName.H5")); err != nil {
		thislog := Log{RequestUrl: "UpdateStaticFile/LoadWeb/h5"}
		thislog.Error("更新静态文件失败", err)
		c.Data(200, "text/html", []byte(err.Error()))
		c.Abort()
		return
	}
	H5FlieMap = H5FlieMapNew // 然后直接覆盖原来的map
	c.Data(200, "text/html", []byte("更新成功"))
	c.Abort()
}
