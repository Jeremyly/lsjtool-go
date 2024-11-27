package timer

import (
	"src/app/model"
	g "src/global"
	"src/utils/base"
	"time"
)

// 数据迁移与备份相关定时任务
func MigrateDataTicker() {
	g.ZeroStamp = base.GetTodayTimeStamp()                                     // 获取当天的0点时间戳
	nextFiveStamp := g.ZeroStamp + 86400 + 18000                               // 加一天的秒数，再加5个小时。得到第二天的5点
	fiveTimer := time.NewTimer(time.Until(time.Unix(int64(nextFiveStamp), 0))) //生成一个明天五点执行的定时器
	TimerList = append(TimerList, fiveTimer)                                   // 添加到队列，方便结束程序时退出定时器
	go func() {
		for {
			select {
			case <-fiveTimer.C:
				migrateImgTaskData() // 迁移图片任务数据
				migrateOrderData()   // 迁移订单数据
				nextFiveStamp += 86400
				fiveTimer.Reset(time.Until(time.Unix(int64(nextFiveStamp), 0)))
			case <-StopChan:
				return // 退出协程
			}
		}
	}()
}

// 将创建时间超过24小时的img_task表数据迁移到img_task_backup表中,并删除原表中的数据
func migrateImgTaskData() {
	thisLog := g.Log{RequestUrl: "migrateImgTaskData"}
	var imgTasks []model.ImgTask
	// 查询img_task表中创建时间超过24小时的数据
	if err := g.Mysql.Where("updated_at < ?", time.Now().Add(-24*time.Hour).UnixMilli()).Find(&imgTasks).Error; err != nil {
		thisLog.Error("查询img_task表数据失败", err)
		return
	}
	if len(imgTasks) == 0 {
		return
	}
	// 将查询到的数据插入到img_task_backup表中
	if err := g.Mysql.Table("img_task_backup").Create(imgTasks).Error; err != nil {
		thisLog.Error("插入数据到img_task_backup表失败", err)
		return
	}
	// 删除img_task表中的数据
	if err := g.Mysql.Delete(imgTasks).Error; err != nil {
		thisLog.Error("删除img_task表中的数据失败", err)
		thisLog.Info("需删除数据：", imgTasks)
		return
	}
}

// 将创建时间超过24小时的Order表数据迁移到Order_backup表中,并删除原表中的数据
func migrateOrderData() {
	thisLog := g.Log{RequestUrl: "migrateOrderData"}
	var orderList []model.Order
	// 查询img_task表中创建时间超过24小时的数据
	if err := g.Mysql.Where("created_at < ?", time.Now().Add(-24*time.Hour).UnixMilli()).Find(&orderList).Error; err != nil {
		thisLog.Error("查询order表数据失败", err)
		return
	}
	if len(orderList) == 0 {
		return
	}
	// 将查询到的数据插入到img_task_backup表中
	if err := g.Mysql.Table("order_backup").Create(orderList).Error; err != nil {
		thisLog.Error("插入数据到order_backup表失败", err)
		return
	}
	// 删除img_task表中的数据
	if err := g.Mysql.Delete(orderList).Error; err != nil {
		thisLog.Error("删除order表中的数据失败", err)
		thisLog.Info("需删除数据：", orderList)
		return
	}
}
