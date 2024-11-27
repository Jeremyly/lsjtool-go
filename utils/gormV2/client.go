/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-12 19:48:59
 * @Description:
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package gormV2

import (
	"fmt"
	"time"

	g "src/global"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// 获取一个 mysql 客户端
func GetOneMysqlClient() (*gorm.DB, error) {
	return GetSqlDriver()
}

// 获取数据库驱动, 可以通过options 动态参数连接任意多个数据库
func GetSqlDriver(dbConf ...ConfigParams) (*gorm.DB, error) {

	var dbDialector gorm.Dialector
	if val, err := getDbDialector(); err != nil {
		g.ZapLog.Error("gorm dialector 初始化失败,dbType:mysql", zap.Error(err))
	} else {
		dbDialector = val
	}
	gormDb, err := gorm.Open(dbDialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",   // table name prefix, table for `User` would be `t_users`
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
		PrepareStmt: true,
		Logger:      redefineLog(), //拦截、接管 gorm v2 自带日志
	})
	if err != nil {
		//gorm 数据库驱动初始化失败
		return nil, err
	}

	// 查询没有数据，屏蔽 gorm v2 包中会爆出的错误
	// https://github.com/go-gorm/gorm/issues/3789  此 issue 所反映的问题就是我们本次解决掉的
	_ = gormDb.Callback().Query().Before("gorm:query").Register("disable_raise_record_not_found", MaskNotDataError)

	// https://github.com/go-gorm/gorm/issues/4838
	// _ = gormDb.Callback().Create().Before("gorm:before_create").Register("CreateBeforeHook", CreateBeforeHook)
	// // 为了完美支持gorm的一系列回调函数
	// _ = gormDb.Callback().Update().Before("gorm:before_update").Register("UpdateBeforeHook", UpdateBeforeHook)
	gormDb.Callback().Update().Before("gorm:before_update").Register("UpdateBeforeHook", UpdateBeforeHook)
	// 为主连接设置连接池(43行返回的数据库驱动指针)
	if rawDb, err := gormDb.DB(); err != nil {
		return nil, err
	} else {
		rawDb.SetConnMaxIdleTime(time.Second * 30)
		rawDb.SetConnMaxLifetime(g.VP.GetDuration("Mysql.SetConnMaxLifetime") * time.Second)
		rawDb.SetMaxIdleConns(g.VP.GetInt("Mysql.SetMaxIdleConns"))
		rawDb.SetMaxOpenConns(g.VP.GetInt("Mysql.SetMaxOpenConns"))
		// 调试模式下，则输出info级别的日志，以便调试，非调试模式则输出设置的日志级别
		if g.VP.GetBool("AppDebug") {
			return gormDb.Debug(), nil
		} else {
			return gormDb, nil
		}
	}
}

// 获取一个数据库方言(Dialector),通俗的说就是根据不同的连接参数，获取具体的一类数据库的连接指针
func getDbDialector() (gorm.Dialector, error) {
	Host := g.VP.GetString("Mysql.Host")
	DataBase := g.VP.GetString("Mysql.DataBase")
	Port := g.VP.GetInt("Mysql.Port")
	User := g.VP.GetString("Mysql.User")
	Pass := g.VP.GetString("Mysql.Pass")
	Charset := g.VP.GetString("Mysql.Charset")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", User, Pass, Host, Port, DataBase, Charset)
	dbDialector := mysql.Open(dsn)
	return dbDialector, nil
}

// 创建自定义日志模块，对 gorm 日志进行拦截、
func redefineLog() gormLog.Interface {
	return createCustomGormLog(
		SetInfoStrFormat("[info] %s\n"),
		SetWarnStrFormat("[warn] %s\n"),
		SetErrStrFormat("[error] %s\n"),
		SetTraceStrFormat("[traceStr] %s [%.3fms] [rows:%v] %s\n"),
		SetTracWarnStrFormat("[traceWarn] %s %s [%.3fms] [rows:%v] %s\n"),
		SetTracErrStrFormat("[traceErr] %s %s [%.3fms] [rows:%v] %s\n"))
}
