/*
 * @Author: lsjweiyi
 * @Date: 2023-07-31 21:35:17
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-15 22:53:21
 * @Description:
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package myFileUtil

import (
	"os"
)

/**
 * @description: 检查文件夹是否存在，如果不存在则尝试创建
 * @param {string} path
 * @return {*}
 */
func PathExistsAndCreat(path string) (bool, error) {
	_, err := os.Stat(path)
	// 表示存在
	if err == nil {
		return true, nil
	}
	// 表示不存在
	if os.IsNotExist(err) {
		// 创建文件夹
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return false, err
		}
		return true, nil // 表示创建成功
	}
	// 表示不确定存不存在
	return false, err
}

/**
 * @description:检查文件或文件夹是否存在
 * @param {string} path
 * @return {*}
 */
func FileOrPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { // 表示存在
		return true, nil
	}
	if os.IsNotExist(err) { // 表示不存在
		return false, nil
	}
	return false, err // 表示不确定存不存在
}

