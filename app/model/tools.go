/*
 * @Author: lsjweiyi
 * @Date: 2023-07-31 21:02:00
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 12:40:19
 * @Description:
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package model

import (
	g "src/global"
)

var ToolsMap = make(map[uint8]Tools)

type Tools struct {
	Id            uint8   `json:"id"`
	Name          string  `json:"name"`
	TokenId       int     `json:"-"`
	SaveDir       string  `json:"-"`
	Price         float64 `json:"price"`
	DiscountPrice float64 `json:"discountPrice"`
}

/**
 * @description: 从数据库加载工具的配置
 * @return {*}
 */
func LoadToolConfig() error {
	ToolsMap = make(map[uint8]Tools)
	toolsList := make([]Tools, 0)
	// 从数据库查询工具信息
	if err := g.Mysql.Table("tools").Find(&toolsList).Error; err != nil {
		return err
	}
	for _, tool := range toolsList {
		ToolsMap[tool.Id] = tool
	}
	return nil
}
