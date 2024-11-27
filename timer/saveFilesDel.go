/*
 * @Author: lsjweiyi
 * @Date: 2024-03-27 22:15:51
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-27 22:45:39
 * @Description: 定时删除保存的文件
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */

package timer

import (
	"os"
	"path/filepath"
	g "src/global"
	"src/utils/base"
	"time"
)

// 数据迁移与备份相关定时任务
func DelSaveFilesTicker() {
	g.ZeroStamp = base.GetTodayTimeStamp()                                     // 获取当天的0点时间戳
	nextFiveStamp := g.ZeroStamp + 86400 + 14400                               // 加一天的秒数，再加5个小时。得到第二天的5点
	fiveTimer := time.NewTimer(time.Until(time.Unix(int64(nextFiveStamp), 0))) //生成一个明天五点执行的定时器
	TimerList = append(TimerList, fiveTimer)                                   // 添加到队列，方便结束程序时退出定时器

	go func() {
		for {
			select {
			case <-fiveTimer.C:
				delSaveFiles()    // 删除数据
				nextFiveStamp += 86400
				fiveTimer.Reset(time.Until(time.Unix(int64(nextFiveStamp), 0)))
			case <-StopChan:
				return // 退出协程
			}
		}
	}()
}



// 删除./save目录及其只目录下最后更新时间超过24小时的文件
func delSaveFiles() {
	thisLog:=g.Log{RequestUrl: "delSaveFiles"}
	// 获取当前时间
	currentTime := time.Now()
	// 获取24小时前的时间
	twentyFourHoursAgo := currentTime.Add(-24 * time.Hour)

	// 遍历./save目录及其子目录下的文件
	err := filepath.Walk("./save", func(path string, info os.FileInfo, err error) error {
		// 检查文件是否为普通文件且最后更新时间超过24小时
		if !info.IsDir() && info.ModTime().Before(twentyFourHoursAgo) {
		    // 删除文件
		    err := os.Remove(path)
		    if err != nil {
		        return err
		    }
		}
		return nil
	})

	if err != nil {
		thisLog.Error("删除文件失败",err)
	}
}
