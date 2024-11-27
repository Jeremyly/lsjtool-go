/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:02
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-08 22:23:08
 * @Description:
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"src/app/model"
	"src/app/service/payService"
	"src/global"
	g "src/global"
	"src/router"
	"src/timer"
	myFileUtil "src/utils/file"
	"src/utils/gormV2"
	"src/utils/mylog"
	snow "src/utils/snowId"
	"src/utils/ymlConfig"
	"src/utils/zapFactory"
	"syscall"
	"time"
)

// 这里可以存放后端路由（例如后台管理系统）
func main() {
	programInit()
	thisLog := g.Log{RequestUrl: "main"}

	// 初始化IP管理模块
	var err error
	g.IP, err = g.InitIpVisit(127, 60, 5)
	if err != nil {
		thisLog.Error(err.Error())
		os.Exit(-1)
	}
	// 加载网站静态资源
	if err = g.LoadWebStatic(); err != nil {
		thisLog.Error("加载网站静态资源失败", err)
		os.Exit(-1)
	}
	router := router.InitRouter() // 初始化路由

	server := http.Server{
		Addr:    g.VP.GetString("HttpServer.Port"),
		Handler: router,
	}
	// 加载支付客户端
	if !LoadPayClient() {
		os.Exit(-1)
	}
	var server80 *http.Server
	go func() {
		if !g.VP.GetBool("AppDebug") || g.VP.GetString("Mysql.DataBase") == global.VP.GetString("DataBaseTest") { // 生产环境， 且不是测试库
			isExist, err := myFileUtil.FileOrPathExists(g.VP.GetString("CA.Crt"))
			if !isExist {
				thisLog.Error("读取SSL证书失败", err)
				os.Exit(-1)
			}
			isExist, err = myFileUtil.FileOrPathExists(g.VP.GetString("CA.Key"))
			if !isExist {
				thisLog.Error("读取SSL证书失败:" + err.Error())
				os.Exit(-1)
			}
			// 配置HTTPS,仅支持TLS1.1+
			server.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS11,
				MaxVersion: tls.VersionTLS13,
			}
			server80 = startHttp80Server()                                   // 启动http80服务,用于重定向到https
			server.ErrorLog = log.New(&mylog.ServerLog{}, "", log.LstdFlags) // 自定义log，否则他一直打印一些无关紧要的错误
			server.ListenAndServeTLS(g.VP.GetString("CA.Crt"), g.VP.GetString("CA.Key"))
		} else {
			// 开发环境
			server.ListenAndServe()
		}
	}()
	timer.StartTimer()                    // 启动业务定时器
	listenProgramExist(&server, server80) //  用于系统信号的监听
}

/**
 * @description: 程序启动初始化
 * @return {*}
 */
func programInit() {

	//2.检查配置文件以及日志目录等非编译性的必要条件
	checkRequiredFolders()
	var err error
	if g.VP, err = ymlConfig.CreateYamlFactory(); err != nil {
		log.Fatal("读取配置文件失败:" + err.Error())
		os.Exit(-1)
	}

	// 5.初始化全局日志句柄，并载入日志钩子处理函数
	g.ZapLog = zapFactory.CreateZapFactory(zapFactory.ZapLogHandler)

	// 6.根据配置初始化 gorm mysql 全局 *gorm.Db
	if g.Mysql, err = gormV2.GetOneMysqlClient(); err != nil {
		log.Fatal("Gorm 数据库驱动、连接初始化失败:" + err.Error())
		os.Exit(-1)
	}

	snow.CreateIdgen() // 初始化雪花id的参数

	if err := model.LoadToolConfig(); err != nil {
		log.Fatal("加载工具配置失败:" + err.Error())
		os.Exit(-1)
	}
}

/**
 * @description: 监听系统退出信号，并做一些处理
 * @return {*}
 */
func listenProgramExist(srv *http.Server, srv80 *http.Server) {
	thisLog := g.Log{RequestUrl: "listenProgramExist"}

	signal.Notify(g.ExitSignal, os.Interrupt, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM) // 监听可能的退出信号

	received := <-g.ExitSignal //接收信号管道中的值
	thisLog.Info("收到信号，进程开始退出", received.String())
	// 先退出http服务，这样外部请求就不会再进来
	thisLog.Info("开始退出服务", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if srv80 != nil {
		if err := srv80.Shutdown(ctx); err != nil {
			thisLog.Error("80服务退出异常", err)
			os.Exit(-1)
		}
	}
	if err := srv.Shutdown(ctx); err != nil {
		thisLog.Error("服务退出异常", err)
		os.Exit(-1)
	}

	thisLog.Info("http服务正常退出", nil)
	// 关闭IP管理模块
	thisLog.Info("关闭IP管理模块成功", nil)
	g.SaveBanIP() // 关闭前保存IP黑名单
	thisLog.Info("保存IP黑名单成功", nil)

	close(g.ExitSignal) // 关闭信号通道

	// 关闭定时器和定时任务的协程
	timer.Stop()
	thisLog.Info("关闭定时器成功", nil)

	thisLog.Info("程序正常退出", nil)
	os.Exit(15)
}

/**
 * @description: 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
 * @return {*}
 */
func checkRequiredFolders() {
	//1.检查配置文件是否存在
	if _, err := os.Stat(g.BasePath + "/config/config.yml"); err != nil {
		log.Fatal("config.yml 配置文件不存在" + err.Error())
	}
}

// 再起一个监听http 80端口的sever,因为好多人他在浏览器输入时，居然选择使用http来请求，导致反问不了，所以需要再起一个服务监听80端口，然后由他重定向到https的443端口
func startHttp80Server() *http.Server {
	if g.VP.GetString("Mysql.DataBase") == global.VP.GetString("DataBaseTest") { // 测试环境不启动
		return nil
	}
	thisLog := g.Log{RequestUrl: "startHttp80Server"}
	server80 := http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.Path, http.StatusMovedPermanently)
		}),
	}
	go func() {
		if err := server80.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				thisLog.Info("http 80端口服务正常退出", nil)
				return
			}
			thisLog.Error("启动80端口监听失败", err)
			os.Exit(-1)
		}
	}()
	return &server80
}

func LoadPayClient() bool {
	return payService.NewAliPayClient() && payService.NewWechatClient()
}
