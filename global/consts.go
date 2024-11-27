/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-01 13:38:24
 * @Description: 保存一些常量
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package global

// 这里定义的常量，一般是具有错误代码+错误说明组成，一般用于接口返回
const (
	// 表单验证器前缀
	ValidatorParamsCheckFailCode int    = -400
	ValidatorParamsCheckFailMsg  string = "参数校验失败"

	ExecuteErrorCode int = -1 // 程序正常执行但未获得预期结果

	StatusOkCode int    = 200 // 请求成功
	StatusOkMsg  string = "Success"

	// CURD 常用业务状态码
	CurdCreatFailCode  int    = -201
	CurdCreatFailMsg   string = "数据库插入失败"
	CurdUpdateFailCode int    = -202
	CurdUpdateFailMsg  string = "数据库更新失败"
	CurdDeleteFailCode int    = -203
	CurdDeleteFailMsg  string = "数据库删除失败"
	CurdSelectFailCode int    = -204
	CurdSelectFailMsg  string = "数据库查询失败"

	// 未知异常
	UnknownErrorCode int    = -500
	UnknownErrorMsg  string = "系统发生未知异常，请联系系统管理员"

	// 访问受限
	AccessForbiddenCode int    = -403
	AccessForbiddenMsg  string = "禁止访问"

	// 访问成功
	AccessSuccessCode int = 200

	// mysql context 绑定的log对象的key 名
	MysqlLogKeyName = "myLog"
	LogIdKey        = "log-id"
	TaskIdKey       = "task-id"
	ToolIdKey       = "tool-id"

	DateFormat = "2006-01-02 15:04:05" //  设置全局日期时间格式

	// 每个ip每个周期允许访问的次数
	IPPerAllow = 50
)
