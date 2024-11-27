/*
 * @Author: lsjweiyi 759209794@qq.com
 * @Date: 2023-11-13 19:23:17
 * @LastEditors: lsjweiyi
 * @LastEditTime: 2024-05-09 21:15:39
 * @FilePath: \ip-black\blackList\onlyIp\ipStru.go
 * @Description: ip管理模块的实现
 */
package global

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"time"
	"unsafe"
)

var (
	cnIpMaskList = make([]*net.IPNet, 0)
)

// 访问ip管理
type ipVisitS struct {
	// 黑名单列表，正好对应ip地址的四层，以下标计数，-1表示永久封禁
	// int8 既可计数，又可表达状态其中，每访问一次，次数加1，当次数累计达到多少时，则可以标记状态
	// 状态的表示使用负数，因为取值范围是[-128,127],所以规定：-128表示永久封禁用户，剩余的负数表示临时封禁的时长，单位分钟
	// 有定时任务每cycleSecond循环检查一次，当取值范围为[-127,0)时，加1，表示封禁时长减一分钟。当取值范围为(0,127]时，归零，表示封禁时长减一分钟
	// 应对场景，防止恶意高频访问。不需要应对非常精确的场景，优点，简单，节省内存，速度非常快，比基于redis的更快，因为没有通信消耗。
	// 方案弊端：1. 按分钟计算，每分钟都得做一次循环，占用性能
	// 2. 封禁时长不可超过127分钟，超过127分钟，将无法计算，只能升级为永封
	// 3. 访问次数每分钟上限限制在127次，再往上无法计算了const
	// 4. 无法记录每次访问的时间，每次巡查只能直接清零
	// 5. 无法记录封禁原因
	// 6. 无法精准的执行封禁时间，因为每次循环本身就需要执行时间，这个执行时间是不计算进封禁时间的
	IpList            [256]*[256]*[256]*[256]int8
	limit             int8  // 设置的每cycleSecond的访问上限
	cycleSecond       int32 // 每次循环检查的间隔时长，单位秒
	visitLimitBanTime int8  // 访问超限的封禁时长，单位分钟
}

/**
 * @description: 初始化
 * @param {int8} limit 设置的每cycleSecond的访问上限
 * @param {int32} cycleSecond 每次循环检查的间隔时长，单位秒
 * @param {int8} visitLimitBanTime 访问超限的封禁时长，单位分钟
 * @return {*ipVisitS} 对象
 * @return {error} 错误
 */
func InitIpVisit(limit int8, cycleSecond int32, visitLimitBanTime int8) (*ipVisitS, error) {
	if !(limit >= 0 && limit <= 127) {
		return nil, errors.New("limit的取值范围是 [0,127]")
	} else if !(cycleSecond >= 1 && cycleSecond <= 3600) {
		return nil, errors.New("cycleSecond的取值范围是 [1,3600]秒")
	} else if !(visitLimitBanTime >= 1 && visitLimitBanTime <= 127) {
		return nil, errors.New("visitLimitBanTime的取值范围是 [1,127]分钟")
	}

	ipVisit := &ipVisitS{
		limit:             limit,
		cycleSecond:       cycleSecond,
		visitLimitBanTime: visitLimitBanTime,
	}

	ipVisit.loadBanIP() // 加载IP黑名单列表

	if err := loadIpMask(); err != nil { // 加载ip白名单
		return nil, err
	}
	return ipVisit, nil
}

/*
*
  - @description: 将一个ip 字符串添加到黑名单里，单线程下添加5千万ip耗时约20秒
  - @param {string} ipStr 形如127.0.0.1
  - @return {*}
    0: 输入的ip格式错误
    (0,128): cycleSecond时间内访问次数
    (-128,0):封禁时长
    -128:    永封
*/
func (i *ipVisitS) Add(ipStr string) int8 {
	ipByte, errCode := ip2byte(ipStr)
	if errCode != 0 {
		return 0
	}
	// 20240804 暂时解除国外ip限制
	// 只允许国内ip
	// if !ipIn(ipByte) {
	// 	return 0
	// }
	thisIP := i.nil2Create(ipByte)
	// 前面都是初始化的校验，下面才是记录ip
	if thisIP[ipByte[3]] == -128 {
		return -128
	} else if thisIP[ipByte[3]] < 0 { // 小于零说明已经被封禁
		thisIP[ipByte[3]]-- // 封禁时长+1
		return thisIP[ipByte[3]]
	} else if thisIP[ipByte[3]] >= 0 && thisIP[ipByte[3]] < i.limit {
		thisIP[ipByte[3]]++ // 访问次数加1
	} else if thisIP[ipByte[3]] == i.limit { // 访问达到上限，封禁一定时间
		thisIP[ipByte[3]] = -(i.visitLimitBanTime)
	}
	return thisIP[ipByte[3]]
}

/*
*
  - @description: 减少封禁时长
  - @param {string} ipStr 封禁的IP地址
  - @return {int8}     0: 输入的ip格式错误
    (0,128): cycleSecond时间内访问次数
    (-128,0):封禁时长
    -128:    永封
*/
func (i *ipVisitS) Reduce(ipStr string) int8 {
	ipByte, errCode := ip2byte(ipStr)
	if errCode != 0 {
		return 0
	}
	thisIP := i.nil2Create(ipByte)
	thisIP[ipByte[3]]--
	return thisIP[ipByte[3]]
}

/*
*
  - @description: 增加封禁时长，如果还未被封禁，则等于封禁时长，如果已被封禁，则增加banTime
  - @param {string} ipStr 封禁的IP地址
  - @param {int8} banTime 用负数表示，数值表示增加的时长，单位分钟，-128表示永封
  - @return {int8}     0: 输入的ip格式错误
    (0,128): cycleSecond时间内访问次数
    (-128,0):封禁时长
    -128:    永封
*/
func (i *ipVisitS) AddBanTime(ipStr string, banTime int8) int8 {
	ip2byte(ipStr)
	ipByteList, errCode := ip2byte(ipStr)
	if errCode != 0 {
		return 0
	}
	return i.addBanTime(ipByteList, banTime)
}

/*
*
  - @description:同AddBanTime一样，入参形式不同, 增加封禁时长，如果还未被封禁，则等于封禁时长，如果已被封禁，则增加banTime
  - @param {string} ipStr 封禁的IP地址
  - @param {int8} banTime 用负数表示，数值表示增加的时长，单位分钟，-128表示永封
  - @return {int8}     0: 输入的ip格式错误
    (0,128): cycleSecond时间内访问次数
    (-128,0):封禁时长
    -128:    永封
*/
func (i *ipVisitS) AddBanTimeByByte(ip []byte, banTime int8) int8 {
	if len(ip) != 4 {
		return 0
	}
	return i.addBanTime(ip, banTime)
}

/*
*
  - @description: 增加封禁时长，如果还未被封禁，则等于封禁时长，如果已被封禁，则增加banTime
  - @param {string} ipStr 封禁的IP地址
  - @param {int8} banTime 用负数表示，数值表示增加的时长，单位分钟，-128表示永封
  - @return {int8}     0: 输入的ip格式错误
    (0,128): cycleSecond时间内访问次数
    (-128,0):封禁时长
    -128:    永封
*/
func (i *ipVisitS) addBanTime(ip []byte, banTime int8) int8 {
	// 不得输入大于等于0的数
	if banTime >= 0 {
		return 0
	}
	thisIP := i.nil2Create(ip)
	// 前面都是初始化的校验，下面才是记录ip
	if banTime == -128 { // 表示永封
		thisIP[ip[3]] = banTime
	} else {
		if thisIP[ip[3]] >= 0 { // 大于等于零说明还未被封禁
			thisIP[ip[3]] = banTime
		} else if thisIP[ip[3]] < 0 { // 小于零需要判断它再加上banTime会不会达到-128
			if thisIP[ip[3]]+banTime > -128 {
				thisIP[ip[3]] += banTime
			} else {
				thisIP[ip[3]] = -128 // 如果会，使其等于-128
			}
		}
	}
	return thisIP[ip[3]]
}

/*
*
  - @description: 判断一个ip是否被封禁，如果被封禁，则返回封禁时长，否则返回访问次数
  - @param {string} ipStr
  - @return {int8} 第一个返回值的含义：0：表示该ip未记录；>0:表示该ip的访问次数；<0:表示该ip的封禁时长；-128：表示该ip永久封禁
    {bool} 第二个返回值的含义：false:输入的ip有误,true:输入的ip正确
*/
func (i *ipVisitS) IsBan(ipStr string) (int8, bool) {

	ipByteList, errCode := ip2byte(ipStr)
	if errCode != 0 {
		return 0, false
	}
	if i.IpList[ipByteList[0]] == nil || i.IpList[ipByteList[0]][ipByteList[1]] == nil || i.IpList[ipByteList[0]][ipByteList[1]][ipByteList[2]] == nil {
		return 0, true
	}
	return i.IpList[ipByteList[0]][ipByteList[1]][ipByteList[2]][ipByteList[3]], true
}

/**
 * @description: 检查ip是否为空值，如果是，则创建该ip,最后一位的值不设置，需要后续代码自己设置值
 * @param {[]int} []int ip数组
 * @return {*}
 */
func (i *ipVisitS) nil2Create(ipIntList []byte) *[256]int8 {
	// 这里是ip的第二层，为空时就初始化
	if i.IpList[ipIntList[0]] == nil {
		var a [256]*[256]*[256]int8 // 每个位置都对应256个数字，直接初始化256个位置
		var b [256]*[256]int8
		var c [256]int8

		a[ipIntList[1]] = &b
		b[ipIntList[2]] = &c

		i.IpList[ipIntList[0]] = &a
	} else if i.IpList[ipIntList[0]][ipIntList[1]] == nil { // 这里是ip的第三层，为空时就初始化
		// var a [256]*[4]int64
		var b [256]*[256]int8 // 每个位置都对应256个数字，直接初始化256个位置
		var c [256]int8
		b[ipIntList[2]] = &c
		i.IpList[ipIntList[0]][ipIntList[1]] = &b
	} else if i.IpList[ipIntList[0]][ipIntList[1]][ipIntList[2]] == nil { // 这里是ip的第四层，为空时就初始化
		// int64就是64bit，第N个bit就可以代表第N个数字，256/4就是4，将[]int64看成是一个连续的存储空间，所以长度为4的int64数组就可以表示256位数
		var c [256]int8
		i.IpList[ipIntList[0]][ipIntList[1]][ipIntList[2]] = &c
	}
	return i.IpList[ipIntList[0]][ipIntList[1]][ipIntList[2]]
}

// 获取黑名单列表长度
func (i *ipVisitS) GetLen() (listLen int) {
	listLen, _ = i.get(1)
	return
}

// 获取被永久封禁的ip
func (i *ipVisitS) GetPermanentBan() (ipList []byte) {
	_, ipList = i.get(2)
	return
}

// 获取被永久封禁的ip的string形式
func (i *ipVisitS) GetPermanentBanString() []string {
	_, ipListByte := i.get(2)
	ipList := make([]string, len(ipListByte)/4)
	// 解压后，添加到ip列表
	for i := 0; i < len(ipList); i++ {
		ipList[i] = fmt.Sprintf("%d.%d.%d.%d", ipListByte[i*4], ipListByte[i*4+1], ipListByte[i*4+2], ipListByte[i*4+3])
	}
	return ipList
}

/**
 * @description: 获取所有的ip，也就是层层判断数组的哪些值不为空，此时的下标就表示一个值，这种方式下还顺便做了升序排序
 * @return {*}
 */
func (i *ipVisitS) GetAll() (ipList []byte) {
	_, ipList = i.get(3)
	return
}

// 删除一个ip的访问数据 ,-2表示输入的ip有误，-1表示无该ip记录，0 表示删除成功
func (i *ipVisitS) DeleteIp(ipStr string) int {
	ipByte, errCode := ip2byte(ipStr)
	if errCode != 0 {
		return -2
	}
	if i.IpList[ipByte[0]] == nil || i.IpList[ipByte[0]][ipByte[1]] == nil || i.IpList[ipByte[0]][ipByte[1]][ipByte[2]] == nil {
		return -1
	} else if i.IpList[ipByte[0]][ipByte[1]][ipByte[2]][ipByte[3]] == 0 {
		return -1
	}
	i.IpList[ipByte[0]][ipByte[1]][ipByte[2]][ipByte[3]] = 0
	return 0
}

/**
 * @description: 获取ip列表长度或者ip列表
 * @param {int} getType 1：获取长度；2：获取ip永久封禁列表；3：获取完整的ip列表
 * @return {*}
 */
func (ip *ipVisitS) get(getType int) (listLen int, ipStrList []byte) {
	for i := 0; i < 256; i++ {
		if ip.IpList[i] == nil {
			continue
		}
		for j := 0; j < 256; j++ {
			if ip.IpList[i][j] == nil {
				continue
			}
			for k := 0; k < 256; k++ {
				if ip.IpList[i][j][k] == nil {
					continue
				}
				for l := 0; l < 256; l++ {
					if ip.IpList[i][j][k][l] == 0 {
						continue
					}
					// 仅统计数量
					if getType == 1 {
						listLen++
					} else if getType == 2 {
						if ip.IpList[i][j][k][l] == -128 {
							// 生成ip
							ipStrList = append(ipStrList, byte(i), byte(j), byte(k), byte(l))
						}
					} else {
						// 生成ip
						ipStrList = append(ipStrList, byte(i), byte(j), byte(k), byte(l))
					}
				}

			}
		}
	}
	return
}

/**
 * @description: 获取黑名单当前所占的内存
 * @return {*} 字节数
 */
func (ip *ipVisitS) GetSizeOf() uintptr {
	var size uintptr
	size = unsafe.Sizeof(ip.IpList) // 只能获取第一层的数组的内存，后续关联的指针所指向的内存是无法获取的
	// 后续层层查询占用的内存，相加
	for i := 0; i < 256; i++ {
		size += unsafe.Sizeof(ip.IpList[i])
		if ip.IpList[i] == nil {
			continue
		}
		for j := 0; j < 256; j++ {
			size += unsafe.Sizeof(ip.IpList[i][j])
			if ip.IpList[i][j] == nil {
				continue
			}
			for k := 0; k < 256; k++ {
				size += unsafe.Sizeof(ip.IpList[i][j][k])
				if ip.IpList[i][j][k] == nil {
					continue
				}
				size += unsafe.Sizeof(ip.IpList[i][j][k][0]) * 256 // 每一位大小都一样的，所以乘以256即可
			}
		}
	}
	return size
}

// 定时任务,每cycleSecond循环检查一次，当取值范围为[-127,0)时，加1，表示封禁时长减一分钟。当取值范围为(0,127]时，归零，表示封禁时长减一分钟
/**
 * @description: 定时任务,每cycleSecond循环检查一次，当取值范围为[-127,0)时，加1，表示封禁时长减一分钟。当取值范围为(0,127]时，归零，表示访问次数清零。且将不再记录ip段重置为空值，防止一直占用内存
 * @param {chanbool} stopChan 程序外部的输入信号，用于告知协程的无限循环该终止了。定时器stop并不会关闭他们所在的协程，需要额外使用StopChan发出关闭协程的信号
 * @return {*} *time.Ticker
 */
func (ip *ipVisitS) CheckIPList(stopChan chan bool) *time.Ticker {
	ticker := time.NewTicker(time.Second * time.Duration(ip.cycleSecond))
	go func() {
		for {
			select {
			case <-ticker.C:
				// 以下变量记录每一层是否有记录ip,如果都没有记录ip,，则将引用置空，防止一直占用内存
				second := 0
				third := 0
				fourth := 0
				for i := 0; i < 256; i++ {
					if ip.IpList[i] == nil {
						continue
					}
					second = 0 // 进入该段的循环前，清空前面记录的次数
					for j := 0; j < 256; j++ {
						if ip.IpList[i][j] == nil {
							continue
						}
						third = 0 // 进入该段的循环前，清空前面记录的次数
						for k := 0; k < 256; k++ {
							if ip.IpList[i][j][k] == nil {
								continue
							}
							fourth = 0 // 进入该段的循环前，清空前面记录的次数
							for l := 0; l < 256; l++ {
								if ip.IpList[i][j][k][l] == 0 {
									continue
								}
								// 表明该段内记录了ip
								fourth++
								third++
								second++
								// 永封的不处理
								if ip.IpList[i][j][k][l] == -128 {
									continue
								} else if ip.IpList[i][j][k][l] > 0 {
									// 每次减20次,但是要防止出现小于0的情况
									ip.IpList[i][j][k][l] = int8(math.Max(float64(ip.IpList[i][j][k][l])-IPPerAllow, 0))
								} else if ip.IpList[i][j][k][l] < 0 {
									ip.IpList[i][j][k][l]++
								}
							}
							// 如果第四层循环结束还是为0，表明该段没有记录ip，可以置空回收，当然，访问次数或封禁时间刚恢复到0的不包含在内
							if fourth == 0 {
								ip.IpList[i][j][k] = nil
							}
						}
						// 如果第三层循环结束还是为0，表明该段没有记录ip
						if third == 0 {
							ip.IpList[i][j] = nil
						}
					}
					// 如果第二层循环结束还是为0，表明该段没有记录ip
					if second == 0 {
						ip.IpList[i] = nil
					}
				}
			case <-stopChan:
				return // 退出协程
			}
		}
	}()
	return ticker
}

func SaveBanIP() {
	banIp := IP.GetPermanentBan()
	thisLog := Log{RequestUrl: "SaveBanIP"}
	// 压缩
	gzipBuff, err := os.Create(VP.GetString("OtherFile.BanIP"))
	if err != nil {
		thisLog.Error("创建文件失败", err)
		return
	}
	// 写入
	gzipW := gzip.NewWriter(gzipBuff)
	_, err = gzipW.Write(banIp)
	if err != nil {
		thisLog.Error("写入数据失败", err)
		return
	}
	gzipW.Close()
	gzipBuff.Close()
}

func (ip *ipVisitS) loadBanIP() {
	thisLog := Log{RequestUrl: "loadBanIP"}
	// 读取
	gzipBuff, err := os.Open(VP.GetString("OtherFile.BanIP"))
	// 如果文件不存在,直接退出
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		thisLog.Error("打开文件失败", err)
		return
	}
	defer gzipBuff.Close()
	gzipR, err := gzip.NewReader(gzipBuff)
	if err != nil {
		thisLog.Error("初始化解压失败", err)
		return
	}
	defer gzipR.Close()
	// 解压
	banIpByte, err := io.ReadAll(gzipR)
	if err != nil {
		thisLog.Error("读取解压后数据失败", err)
		return
	}

	// 解压后，添加到ip列表
	banIp := make([]int32, len(banIpByte)/4)
	for i := 0; i < len(banIp); i++ {
		ip.AddBanTimeByByte(banIpByte[i*4:(i+1)*4], -128)
	}
}

/**
 * @description: 加载ip列表
 * @return {*}
 */
func loadIpMask() error {
	thisLog := Log{RequestUrl: "loadIpMask"}
	ipMask, err := os.Open(VP.GetString("OtherFile.CnIp"))
	// 如果文件不存在,直接退出
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		thisLog.Error("打开文件失败", err)
		return err
	}
	// 创建ip列表
	ipMaskScanner := bufio.NewScanner(ipMask)
	for ipMaskScanner.Scan() {
		_, ipnet, err := net.ParseCIDR(ipMaskScanner.Text())
		if err != nil {
			return err
		}
		cnIpMaskList = append(cnIpMaskList, ipnet)
	}
	return nil
}

// 通过二分查找
func ipIn(ip []byte) bool {
	low, high := 0, len(cnIpMaskList)-1
	mid := 0
	for low <= high {
		mid = low + (high-low)/2
		if cnIpMaskList[mid].Contains(ip) {
			return true
		} else if byteCompare(cnIpMaskList[mid].IP, ip) {
			high = mid - 1
			continue
		} else {
			low = mid + 1
			continue
		}
	}
	return false
}

/**
 * @description: 将ip转为byte数组
 * @param {string} ipStr ip字符串
 * @return {*} ip数组和错误码，-1表示ip格式错误，否则为0
 */
func ip2byte(ip string) ([]byte, int) {
	ipByte := make([]byte, 4)
	j := 0
	for i := 0; i < len(ip); i++ {
		if ip[i] == byte('.') {
			j++
		} else {
			ipByte[j] = ipByte[j]*10 + byte(ip[i]) - '0'
		}
	}
	if j != 3 {
		return nil, -1
	}
	return ipByte, 0
}

// byte 数组比较大小,确保a!=b, 返回true表示a>b，返回false表示a<b
func byteCompare(a []byte, b []byte) bool {
	for i := 0; i < len(a); i++ {
		if a[i] > b[i] {
			return true
		} else if a[i] < b[i] {
			return false
		}
	}
	return true
}
