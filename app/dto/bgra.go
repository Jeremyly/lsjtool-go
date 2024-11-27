/*
 * @Author: lsjweiyi
 * @Date: 2023-06-28 20:33:02
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2023-07-07 21:12:50
 * @Description: 图片四通道的结构体
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package dto

import "gocv.io/x/gocv"

// RGBA四通道
type BRGA struct {
	B int
	G int
	R int
	A int
}

// 返回gocv.Scalar形式的BGRA结构体，且归一化
/**
 * @description: 返回gocv处理时需要的*gocv.Scalar
 * @return {*}
 */
func (b *BRGA) GetOneScalar() *gocv.Scalar {
	return &gocv.Scalar{Val1: float64(b.B) / 255.0, Val2: float64(b.G) / 255.0, Val3: float64(b.R) / 255.0, Val4: float64(b.A) / 255.0}
}
