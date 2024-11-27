/*
 * @Author: lsjweiyi
 * @Date: 2024-03-15 19:36:42
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-04 01:15:55
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package middleware

import (
	g "src/global"
	"src/utils/response"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/yitter/idgenerator-go/idgen"
)

/**
 * @description: 创建新的任务id
 * @return {*} 用户id
 */
func CreateTask() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 暴露该请求头给前端,和logId 一并设置了
		c.Writer.Header().Add("Access-Control-Expose-Headers", g.TaskIdKey)
		taskId := ran6LenString()
		c.Request.Header.Set(g.TaskIdKey, taskId) // 设置request的header
		c.Header(g.TaskIdKey, taskId)             // 设置response的header
		c.Next()
	}
}

/**
 * @description: 校验任务id
 * @return {*} 用户id
 */
func CheckTaskId() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskId := c.Request.Header.Get(g.TaskIdKey)
		if taskId == "" {
			response.Fail(c, g.AccessForbiddenCode, "必须输入任务id")
			return
		}
		if !CheckUserId(taskId) { // 如果是前台传的taskid，那就得校验
			response.Fail(c, g.AccessForbiddenCode, "输入的任务编号有误，请重新输入")
			return
		}
		c.Next()
	}
}

/**
 * @description: 随机生成一个长度为6的字符串，再拼接两位hash校验码。仅包含数字和大写字母
 * @return {*} 长度为8的字符串
 */
func ran6LenString() string {
	result := make([]byte, 6)
	a := idgen.NextId() // 雪花算法获取随机数
	total := 0
	// 求所有位数上数字和i的乘积的和
	copyA := a
	for i := 19; i > 0; i-- {
		total += int(copyA%10) * i
		copyA /= 10
	}
	// 每三位为一段，求每三位和i的乘积和，用total减去该和再取余62
	for i := 6; i > 0; i-- {
		tmpSum := 0
		for j := 0; j < 3; j++ {
			tmpSum += i * int(a%10)
			a /= 10
		}
		// 36是因为大写字母加数字共36个
		tmp := (total - tmpSum) % 36
		// 人为的将前10个定义为数字，所以余数小于10时，表示获得的数字
		if tmp < 10 {
			result[i-1] = byte(tmp + 48) // 数字的ascii码是48~57,而其余数的范围是[0,9]，所以映射关系是temp+48
		} else {
			// 人为的将10以后得数定义为字母，所以余数大于等于10时，表示获得的是字母
			result[i-1] = byte(tmp + 55) // 大写字母的ascii码是65～90,而其余数的范围是[10,35]，所以映射关系是temp-10+65
		}
	}
	resStr := *(*string)(unsafe.Pointer(&result))
	return resStr + createUserHashCode(resStr)
}

/**
 * @description: 简单版生成hashCode功能，6位字符，生成两位hash字符，下面组合顺序随意发挥,虽然算法很简单，但是不暴露代码对方也猜不到算法的
 * @param {string} userId
 * @return {*}
 */
func createUserHashCode(userId string) string {
	//将长度6位的字符串hash成两位字符
	hashCode := make([]byte, 2)
	code1 := int(userId[0]) + int(userId[2]) + int(userId[4]) + int(userId[5])
	code2 := int(userId[0]) + int(userId[1]) + int(userId[3]) + int(userId[4])
	// 36是因为大写字母加数字共36个
	// 人为的将前10个定义为数字，所以余数小于10时，表示获得的数字
	if code1%36 < 10 {
		hashCode[0] = byte(code1%36 + 48) // 数字的ascii码是48~57,而其余数的范围是[0,9]，所以映射关系是temp+48
	} else {
		// 人为的将10以后得数定义为字母，所以余数大于等于10时，表示获得的是字母
		hashCode[0] = byte(code1%36 + 55) // 大写字母的ascii码是65～90,而其余数的范围是[10,35]，所以映射关系是temp-10+65
	}
	if code2%36 < 10 {
		hashCode[1] = byte(code2%36 + 48)
	} else {
		hashCode[1] = byte(code2%36 + 55)
	}
	return string(hashCode)
}

/**
 * @description: 检查请求是否传入userid，有则检验是否正确
 * @param {*gin.Context} c
 * @return {*}
 */
func CheckUserId(taskId string) bool {
	// 检查id格式是否合法
	if len(taskId) != 8 {
		return false
	}
	for _, v := range taskId {
		// 不是数字且不是大写字母
		if !(v >= 48 && v <= 57) && !(v >= 65 && v <= 90) {
			return false
		}
	}
	// 校验hash码，防止用户自定义userId
	hashCode := createUserHashCode(taskId[:6])
	return hashCode == taskId[6:]
}
