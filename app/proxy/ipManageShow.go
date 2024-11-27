/*
 * @Author: lsjweiyi
 * @Date: 2024-03-04 18:49:26
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-25 18:51:31
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package proxy

import (
	"src/global"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetLen(c *gin.Context) {
	c.JSON(200, global.IP.GetLen())
}
func GetPermanentBan(c *gin.Context) {
	c.JSON(200, global.IP.GetPermanentBanString())
}

// func GetAll(c *gin.Context) {
// 	c.JSON(200, global.IP.GetAll())
// }

func GetSizeOf(c *gin.Context) {
	c.JSON(200, global.IP.GetSizeOf())
}

func DeleteIP(c *gin.Context){
	c.JSON(200, global.IP.DeleteIp(c.Query("ip")))
}

func AddBanTime(c *gin.Context) {
	time,err:=strconv.Atoi(c.Query("time"))
	if err!=nil{
	    c.JSON(200, gin.H{"error": "time must be a number"})
	    return
	}
    c.JSON(200, global.IP.AddBanTime(c.Query("ip"), int8(time)))
}
