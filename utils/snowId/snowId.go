/*
 * @Author: lsjweiyi
 * @Date: 2023-02-10 20:44:01
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-03-13 21:02:27
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package snowId

import "github.com/yitter/idgenerator-go/idgen"

// 调用包创建唯一id，具体说明见：https://www.lmlphp.com/user/62248/article/item/625833/
// 使用场景1：日志id，用于追踪日志。2：数据唯一id，代替数据库的自增。
// 算法基于雪花算法，据作者描述，比传统雪花算法好，精力有限，没有细究
// 只需要将该方法在启动程序时调用一遍，即可将参数注册到包中，之后还是要引入包调用方法：idgen.NextId()获取id
func CreateIdgen() {
	var options = idgen.NewIdGeneratorOptions(1) // 单机版机器码，直接编一个上去，原文有介绍如何在集群中生成唯一机器码
	options.SeqBitLength = 12
	idgen.SetIdGenerator(options)
}
