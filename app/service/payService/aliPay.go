package payService

import (
	"context"
	"errors"
	"fmt"
	"os"
	"src/app/model"
	g "src/global"
	"src/utils/base"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

var (
	aliAppid        string
	aliPrivateKey   string
	aliPayPublicKey string
	aliPayClient    *alipay.Client
	aliNotifyUrl    string
)

func NewAliPayClient() bool {
	thisLog := g.Log{RequestUrl: "NewAliPayClient"}
	if !GetAliConstants(thisLog) { // 获取支付宝常量
		return false
	}
	// 初始化支付宝客户端
	// appid：应用ID
	// privateKey：应用私钥，支持PKCS1和PKCS8
	// isProd：是否是正式环境，沙箱环境请选择新版沙箱应用。
	var err error
	aliPayClient, err = alipay.NewClient(aliAppid, aliPrivateKey, true)
	if err != nil {
		thisLog.Error("初始化支付宝客户端失败", err)
		return false
	}
	// 打开Debug开关，输出日志，默认关闭
	aliPayClient.DebugSwitch = gopay.DebugOff

	// 设置支付宝请求 公共参数
	//    注意：具体设置哪些参数，根据不同的方法而不同，此处列举出所有设置参数
	aliNotifyUrl = fmt.Sprintf("%s%s/tools/order/aliPayNotify", g.VP.GetString("URL"), g.VP.GetString("HttpServer.Port"))
	aliPayClient.SetNotifyUrl(aliNotifyUrl) // 设置异步通知URL
	return true

}

func AliH5CreateOrder(ctx *gin.Context, order *model.Order) (*string, error) {

	timeExpire := time.Now().Add(time.Minute * 5) // 设置订单过期时间为5分钟
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", order.Id)
	bm.Set("total_amount", model.ToolsMap[order.ToolId].DiscountPrice)
	bm.Set("subject", model.ToolsMap[order.ToolId].Name)
	bm.Set("product_code", "QUICK_WAP_WAY") // 销售产品码，商家和支付宝签约的产品码。手机网站支付为：QUICK_WAP_WAY
	bm.Set("time_expire", timeExpire.Format("2006-01-02 15:04:05"))

	if order.ToolId == 27 { // 因为付款是跳转新的页面，很多人不知道，然后看到页面为空就不会了，所以得自动跳转回去，并且带上taskId
		bm.Set("quit_url", g.VP.GetString("URL"))
		aliPayClient.SetReturnUrl(fmt.Sprintf("%s/?taskId=%s", g.VP.GetString("URL"), order.TaskId)) // 设置支付成功跳转Url
	} else {
		aliPayClient.ReturnUrl = "" // 其他不设置跳转，不然会丢失页面信息
	}

	payUrl, err := aliPayClient.TradeWapPay(ctx.Request.Context(), bm)

	defer order.Insert()

	if err != nil {
		order.ResultType = 2 // 标记为订单创建失败
		order.EndAt = time.Now().UnixMilli()
		// 处理错误
		order.Log.Error(fmt.Sprintf("创建支付宝H5支付订单%s失败", order.Id), err)
		return nil, errors.New("创建支付宝H5支付订单失败,请重试")
	} else {
		order.CreatedAt = time.Now().UnixMilli()
		order.ResultType = 1 // 标记为支付中
	}
	return &payUrl, nil
}

func AlipayNotify(c *gin.Context) error {
	temp, _ := c.Get("log")
	thisLog, _ := temp.(*g.Log)
	// 解析异步通知的参数
	notifyReq, err := alipay.ParseNotifyToBodyMap(c.Request) // c.Request 是 gin 框架的写法
	if err != nil {
		thisLog.Error("解析支付宝异步通知失败", err)
		return errors.New("解析支付宝异步通知失败")
	}
	ok, err := alipay.VerifySign(aliPayPublicKey, notifyReq)
	if !ok {
		thisLog.Error("验证支付宝签名失败", err)
		return errors.New("验证支付宝签名失败")
	}
	// 处理通知内容
	order := model.InitOrder(c, 0, -1)             // 因为此处是用于查询订单，所以先toolid先填0，等待后续查询结果
	order.Id = notifyReq.GetString("out_trade_no") // 内部订单号
	if err := order.Get(); err != nil {
		return err
	}
	order.TradeNo = notifyReq.GetString("trade_no") // 支付宝订单号

	// 转换订单完成时间
	if notifyReq.GetString("gmt_close") == "" { // 订单未成功时，不会有SuccessTime
		order.EndAt = time.Now().UnixMilli()
	} else {
		tm, err := time.Parse("2006-01-02 15:04:05", notifyReq.GetString("gmt_close"))
		if err != nil {
			thisLog.Error("解析时间字符串失败:", err)
			return err
		}
		fmt.Println("gmt_close", tm)
		order.EndAt = tm.Local().UnixMilli()
	}
	order.TradeType = notifyReq.GetString("body")      // 交易类型
	order.Payer = notifyReq.GetString("buyer_open_id") // 支付者openid
	if notifyReq.GetString("trade_status") == "TRADE_SUCCESS" {
		order.ResultType = 0
		order.Reason = "支付成功"
		base.ToolPay(order.ToolId) // 记录次数
		receiptAoumt, err := strconv.ParseFloat(notifyReq.GetString("receipt_amount"), 64)
		if err != nil {
			thisLog.Error("解析支付宝收到的金额失败", err)
			return err
		}
		base.ToolPayAmount(order.ToolId, int64(receiptAoumt*100)) // 默认单位元，转成分
	} else if notifyReq.GetString("trade_status") == "WAIT_BUYER_PAY" { // 交易创建，等待买家付款。
		order.Reason = "等待买家付款"
		order.ResultType = 5
	} else if notifyReq.GetString("trade_status") == "TRADE_CLOSED" { // 未付款交易超时关闭，或支付完成后全额退款
		if notifyReq.GetString("gmt_refund") != "" { // 有退款时间的，就是退款回调
			if order.ResultType == 8 { // 退款时已有同步消息返回处理，此处无需重复处理，只需要正常回应回调就行
				fmt.Println("退款已处理，无需重复处理")
				return nil
			}
			order.ResultType = 8
			base.ToolRefund(order.ToolId) // 记录退款
			refundAoumt, err := strconv.ParseFloat(notifyReq.GetString("refund_fee"), 64)
			if err != nil {
				thisLog.Error("解析支付宝退款金额失败", err)
				return err
			}
			base.ToolRefundAmount(order.ToolId, int64(refundAoumt*100))
		} else { // 订单超过退款期限后，支付宝还会发起回调通知商户
			order.ResultType = 6
			order.Reason = "已关闭"
		}
	} else if notifyReq.GetString("trade_status") == "TRADE_FINISHED" { // 交易结束，不可退款
		// 当前业务不做处理
	} else { // 未知状态
		order.ResultType = 99
		thisLog.Error("订单 " + order.Id + "未知状态" + notifyReq.GetString("trade_status"))
		fmt.Println(notifyReq)
	}
	if err := order.Update(); err != nil { // 更新订单
		return err
	}
	if order.ResultType == 0 {
		return ToolAddDownTimes(order) // 增加次数
	} else if order.ResultType == 8 {
		return ToolRefundReduceDownTimes(order) // 退款减少次数
	} else {
		return nil
	}
}

// 支付宝主动发起查询订单状态
func AliQueryOrderStatus(order *model.Order) {
	thisLog := g.Log{RequestUrl: "AliQueryOrderStatus"}
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", order.Id)
	ctx := context.Background()
	notifyReq, queryErr := aliPayClient.TradeQuery(ctx, bm)

	// 未付款的订单支付宝直接提示查不到，创建的订单未付款支付宝是提示查不到的。所以只能通过是否返回TRADE_NOT_EXIST来处理
	if notifyReq.Response != nil && notifyReq.Response.SubCode == "ACQ.TRADE_NOT_EXIST" {
		notifyReq.Response.TradeStatus = "TRADE_CANCEL" //  将其定义为超时未付款
	} else {
		if queryErr != nil { // 几乎不会出现这种情况
			thisLog.Error("支付宝查询订单状态失败："+order.Id, queryErr)
			return
		}
	}

	order.TradeNo = notifyReq.Response.TradeNo // 支付宝订单号

	// 转换订单完成时间
	order.EndAt = time.Now().UnixMilli()
	order.Payer = notifyReq.Response.BuyerOpenId // 支付者openid

	if notifyReq.Response.TradeStatus == "TRADE_SUCCESS" {
		order.ResultType = 0
		order.Reason = "支付成功"
		base.ToolPay(order.ToolId)
		amount, err := strconv.ParseFloat(notifyReq.Response.TotalAmount, 64)
		if err != nil {
			thisLog.Error("订单 "+order.Id+"金额转换失败", err)
			return
		}
		base.ToolPayAmount(order.ToolId, int64(amount*100)) // 默认单位元，转成分
	} else if notifyReq.Response.TradeStatus == "WAIT_BUYER_PAY" { // 交易创建，等待买家付款。
		order.Reason = "等待买家付款"
		order.ResultType = 5
	} else if notifyReq.Response.TradeStatus == "TRADE_CANCEL" { // 未付款交易超时取消
		order.ResultType = 3
		order.Reason = "未付款交易超时取消"
	} else if notifyReq.Response.TradeStatus == "TRADE_CLOSED" { // 支付完成后全额退款
		order.ResultType = 6
		order.Reason = "支付完成后全额退款"
	} else if notifyReq.Response.TradeStatus == "TRADE_FINISHED" { // 交易结束，不可退款
		// 当前业务不做处理
	} else { // 未知状态
		order.ResultType = 99
		thisLog.Error("订单 " + order.Id + "未知状态" + notifyReq.Response.TradeStatus)
		fmt.Println("未知状态", notifyReq)
	}
	if err := order.Update(); err != nil { // 更新订单
		return
	}

	if order.ResultType != 0 { // 只有支付成功了才需要增加次数
		return
	}
	// 增加次数
	ToolAddDownTimes(order)
}

// 发起退款
func AliRefund(order *model.Order, reason string) error {
	thisLog := g.Log{RequestUrl: "AliRefund"}

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", order.Id)
	bm.Set("refund_reason", reason)
	bm.Set("refund_amount", strconv.FormatFloat(float64(order.Amount)/100, 'f', 2, 64)) // 退款金额

	order.ResultType = 4 // 标记为转入退款状态
	order.Reason = reason
	if err := order.Update(); err != nil { // 更新订单
		return err
	}

	// 实测发现，退款的同步请求返回比回调还慢，所以，为了避免重复处理退款结果，这里不再处理结果，交由回调处理
	ctx := context.Background()
	aliresp, refundErr := aliPayClient.TradeRefund(ctx, bm)
	if refundErr != nil {
		thisLog.Error("支付宝退款失败", refundErr)
		order.Reason = "支付宝退款失败"
		order.ResultType = 9                   // 标记为退款失败状态
		if err := order.Update(); err != nil { // 更新订单
			return err
		}
		return errors.New("支付宝退款失败,请联系管理员")
	}

	fmt.Println("aliresp", aliresp)
	return nil

}

// 从数据库查询常量
func GetAliConstants(thisLog g.Log) bool {
	aliAppid = model.SelectConstant(g.VP.GetString("AliPay.AppIdName"))
	if aliAppid == "" {
		thisLog.Error("支付宝appid常量未配置")
		return false
	}
	var err error

	// 从txt文件里读取私钥
	aliPrivateKeyPath := g.VP.GetString("AliPay.PrivateKey")
	if aliPrivateKeyPath == "" {
		thisLog.Error("加载支付宝私钥路径失败", err)
		return false
	}
	aliPrivateKeyByte, err := os.ReadFile(aliPrivateKeyPath)
	if err != nil {
		thisLog.Error("加载支付宝私钥失败", err)
		return false
	}
	aliPrivateKey = string(aliPrivateKeyByte)

	// 从txt文件里读取公钥
	aliPayPublicKeyPath := g.VP.GetString("AliPay.PublicKey")
	if aliPayPublicKeyPath == "" {
		thisLog.Error("加载支付宝公钥路径失败", err)
		return false
	}
	aliPayPublicKeyByte, err := os.ReadFile(aliPayPublicKeyPath)
	if err != nil {
		thisLog.Error("加载支付宝公钥失败", err)
		return false
	}
	aliPayPublicKey = string(aliPayPublicKeyByte)
	return true
}
