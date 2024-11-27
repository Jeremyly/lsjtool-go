/*
 * @Author: lsjweiyi
 * @Date: 2023-02-13 20:39:17
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-04-20 12:38:45
 * @Description: 将对应的接口需要的token获取并保存到map中
 * Copyright (c) 2023 by lsjweiyi, All Rights Reserved.
 */
package timer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"src/app/service/outApi"
	g "src/global"
	"src/utils/myHttp"
	"time"

	"go.uber.org/zap"
)

type creatToken struct {
	AppID     int
	ApiKey    string
	SecretKey string
	Token     string
}

type tokenResponseS struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

const (
	tokenUrl string = "https://aip.baidubce.com/oauth/2.0/token?client_id=%s&client_secret=%s&grant_type=client_credentials"
)

/**
 * @description: 利用key获取token，每次启动程序执行一次
 */
func LoadToken(stopChan chan bool) error {

	tokenArray := make([]creatToken, 0)
	// 从数据库查询token
	if err := g.Mysql.Table("token").Select("app_id,api_key,secret_key").Find(&tokenArray).Error; err != nil {
		return err
	}
	if len(tokenArray) == 0 {
		return errors.New("未找到任何api接口信息，无法获取token，请检查app信息！")
	}
	// 启动定时器，定时检查token
	tokenTimer(tokenArray, stopChan, 1)

	// 初始化百度app的参数
	outApi.InitBaiduApp()

	return nil
}

/**
 * @description: 定时器，定时检查token，如果token快到期了，就重新获取token
 * @param {[]creatToken} appIdList api接口的key信息
 * @param {chan bool} stopChan 退出协程的信号
 * @param {int} expiresIn token过期时间
 * @return {*}
 */
func tokenTimer(appIdList []creatToken, stopChan chan bool, expiresIn int) {
	ticker := time.NewTicker(time.Second * time.Duration(expiresIn)) // 启动定时器
	TickerList = append(TickerList, ticker)                          // 添加到队列，方便结束程序时退出定时器
	go func() {
		ctx := context.Background()
		for {
			select {
			case <-ticker.C:
				expiresIn = getToken(ctx, appIdList)
				ticker.Reset(time.Second * time.Duration(expiresIn-60)) // 重新设置定时器,比过期时间早一分钟执行
			case <-stopChan:
				ctx.Done()
				return // 退出协程
			}
		}
	}()
}

/**
 * @description: 向百度请求token
 * @param {context.Context} ctx
 * @param {*creatToken} tokenStruct
 * @return {error}
 */
func getToken(ctx context.Context, tokenStructList []creatToken) int {
	var expiresIn int
	for _, tokenStruct := range tokenStructList {
		tokenResponse := tokenResponseS{}
		if err := myHttp.FormPost(ctx, fmt.Sprintf(tokenUrl, tokenStruct.ApiKey, tokenStruct.SecretKey), "", &tokenResponse); err != nil {
			g.ZapLog.Error("获取百度token时请求失败", zap.Error(err))
			os.Exit(-1)
		} else if tokenResponse.Error != "" {
			g.ZapLog.Error("获取百度token时请求失败", zap.Error(err))
			os.Exit(-1)
		}
		g.BaiduToken[tokenStruct.AppID] = tokenResponse.AccessToken
		expiresIn = tokenResponse.ExpiresIn
	}
	return expiresIn
}
