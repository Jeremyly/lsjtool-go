/*
 * @Author: lsjweiyi
 * @Date: 2024-05-08 21:36:23
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-08 22:15:32
 * @Description:
 * Copyright (c) 2024 by lsjweiyi, All Rights Reserved.
 */
package mylog

import (
	"src/global"
)

var (
	thisLog  = global.Log{RequestUrl: "serverLog"}
	unknownC = "tls: unknown certificate"
	sSLv2    = "SSLv2 handshake received"
	noCipher = "no cipher suite supported by both client and server"
	client   = "client offered only unsupported versions: []"
	eof      = "EOF"
)

type ServerLog struct {
}

func (s *ServerLog) Write(a []byte) (n int, err error) {
	// 屏蔽以上几种反复出现的错误，因为很多爬虫来访问网站，因为ssl证书只绑定了域名，爬虫是直接访问ip的，所以会报错
	if unknownC == string(a[len(a)-len(unknownC)-1:len(a)-1]) &&
		sSLv2 == string(a[len(a)-len(sSLv2)-1:len(a)-1]) &&
		noCipher == string(a[len(a)-len(noCipher)-1:len(a)-1]) &&
		client == string(a[len(a)-len(client)-1:len(a)-1]) &&
		eof == string(a[len(a)-len(eof)-1:len(a)-1]) {

		thisLog.Warn("ServerLog", string(a))
	}
	return len(a), nil
}
