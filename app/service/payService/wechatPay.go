package payService

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"src/app/model"

	g "src/global"
	"src/utils/base"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

var (
	mchID         string          // 商户号
	mchCertNum    string          // 商户证书序列号
	mchAPIv3Key   string          // 商户APIv3密钥
	fansongAppId  string          // 凡松小程序appid
	mchPrivateKey *rsa.PrivateKey // 商户私钥
	// client        *core.Client    // 微信支付客户端
	notifyHandler *notify.Handler // 支付回调handle
	notifyUrl     string
	wechatSvc     *native.NativeApiService
	refundSvc     *refunddomestic.RefundsApiService // 退款服务
)

// 创建微信支付客户端
func NewWechatClient() bool {
	thisLog := g.Log{RequestUrl: "NewWechatClient"}

	// 读取微信的常量，读取不到程序是不能运行的
	if !GetWechatConstants(thisLog) { // 获取微信常量
		return false
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertNum, mchPrivateKey, mchAPIv3Key),
	}
	var err error
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		thisLog.Error("创建微信支付客户端失败", err)
		return false
	}
	// 初始化微信支付API服务
	wechatSvc = &native.NativeApiService{Client: client}
	// 初始化退款API服务
	refundSvc = &refunddomestic.RefundsApiService{Client: client}
	notifyUrl = fmt.Sprintf("%s%s/tools/order/wechatNotify", g.VP.GetString("URL"), g.VP.GetString("HttpServer.Port"))
	return getNoTifyHandler() // 最后获取notifyHandler
}

// 初始化回调控制器
func getNoTifyHandler() bool {
	thisLog := g.Log{RequestUrl: "getNoTifyHandler"}

	ctx := context.Background()
	// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
	err := downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, mchCertNum, mchID, mchAPIv3Key)
	if err != nil {
		// 处理错误
		thisLog.Error("注册下载器失败", err)
		return false
	}
	// 2. 获取商户号对应的微信支付平台证书访问器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(mchID)
	// 3. 使用证书访问器初始化 `notify.Handler`
	notifyHandler = notify.NewNotifyHandler(mchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
	return true
}

// 创建微信native支付
func WechatNativeCreateOrder(ctx *gin.Context, order *model.Order) (*string, error) {
	description := string(order.ToolId) // 订单描述
	amount := native.Amount{            // 订单金额
		Total: &order.Amount,
	}
	timeExpire := time.Now().Add(time.Minute * 5) // 订单有效期
	// 创建订单
	resp, _, err := wechatSvc.Prepay(ctx.Request.Context(),
		native.PrepayRequest{
			Appid:       &fansongAppId,
			Mchid:       &mchID,
			Description: &description,
			OutTradeNo:  &order.Id,
			Amount:      &amount,
			NotifyUrl:   &notifyUrl,
			TimeExpire:  &timeExpire, // 订单有效时间五分钟
		},
	)

	defer order.Insert() // 记录订单

	if err != nil {
		order.ResultType = 2 // 标记为订单创建失败
		order.EndAt = time.Now().UnixMilli()
		// 处理错误
		order.Log.Error(fmt.Sprintf("创建微信native支付订单%s失败", order.Id), err)
		return nil, errors.New("创建微信支付订单失败,请重试")
	} else {
		order.CreatedAt = time.Now().UnixMilli()
		order.ResultType = 1 // 标记为支付中
	}
	return resp.CodeUrl, nil
}

// 支付回调
func WechatNotify(c *gin.Context) error {
	temp, _ := c.Get("log")
	thisLog, _ := temp.(*g.Log)

	transaction := new(payments.Transaction)
	_, err := notifyHandler.ParseNotifyRequest(context.Background(), c.Request, transaction)
	// 如果验签未通过，或者解密失败
	if err != nil {
		thisLog.Error("支付回调验签失败", err)
		return err
	}
	// 处理通知内容
	order := model.InitOrder(c, 0, -1) // 因为此处是用于查询订单，所以先toolid先填0，等待后续查询结果
	order.Id = *transaction.OutTradeNo
	if err := order.Get(); err != nil {
		return err
	}

	order.TradeNo = *transaction.TransactionId // 微信订单号

	// 转换订单完成时间
	if transaction.SuccessTime == nil { // 订单未成功时，不会有SuccessTime
		order.EndAt = time.Now().UnixMilli()
	} else { // 有SuccessTime时，解析时间
		tm, err := time.Parse(time.RFC3339, *transaction.SuccessTime)
		if err != nil {
			thisLog.Error("解析 RFC3339 时间字符串失败:", err)
			return err
		}
		order.EndAt = tm.Local().UnixMilli()
	}

	if transaction.TradeType != nil {
		order.TradeType = *transaction.TradeType // 交易类型
	}

	if transaction.Payer != nil {
		order.Payer = *transaction.Payer.Openid // 支付者openid
	}
	order.Reason = *transaction.TradeStateDesc // 交易状态描述

	if *transaction.TradeState == "SUCCESS" {
		order.ResultType = 0
		base.ToolPay(order.ToolId) // 记录次数
		base.ToolPayAmount(order.ToolId, *transaction.Amount.PayerTotal)
	} else if *transaction.TradeState == "REFUND" { // 转入退款
		order.ResultType = 4
	} else if *transaction.TradeState == "NOTPAY" { // 未支付
		order.ResultType = 5
	} else if *transaction.TradeState == "CLOSED" { // 已关闭
		order.ResultType = 5
	} else if *transaction.TradeState == "PAYERROR" { // 支付失败
		order.ResultType = 99
	} else { // 未知状态
		order.ResultType = 99
	}
	if err := order.Update(); err != nil { // 更新订单
		return err
	}

	if order.ResultType != 0 { // 只有支付成功了才需要增加次数
		return nil
	}
	// 增加次数
	return ToolAddDownTimes(order)
}

// 微信主动发起查询订单状态
func WechatQueryOrderStatus(orderId string) {
	thisLog := g.Log{RequestUrl: "WechatQueryOrderStatus"}
	resp, _, queryErr := wechatSvc.QueryOrderByOutTradeNo(context.Background(),
		native.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(orderId),
			Mchid:      core.String(mchID),
		},
	)

	// 处理通知内容
	order := model.Order{Id: orderId}
	order.DB = g.Mysql // 这里需要手工初始化
	order.Log = &thisLog
	if err := order.Get(); err != nil {
		thisLog.Error("查询订单失败", err)
		return
	}

	if queryErr != nil { // 几乎不会出现这种情况
		thisLog.Error("微信查询订单状态失败："+orderId, queryErr)
		return
	}

	// 转换订单完成时间
	if resp.SuccessTime == nil { // 订单未成功时，不会有SuccessTime
		order.EndAt = time.Now().UnixMilli()
	} else { // 有SuccessTime时，解析时间
		tm, err := time.Parse(time.RFC3339, *resp.SuccessTime)
		if err != nil {
			thisLog.Error("解析 RFC3339 时间字符串失败:", err)
			return
		}
		order.EndAt = tm.Local().UnixMilli()
	}

	if resp.TradeType != nil { // 订单未成功时，不会有TradeType
		order.TradeType = *resp.TradeType // 交易类型
	}

	if resp.Payer != nil { // 订单未成功时，不会有Payer
		order.Payer = *resp.Payer.Openid // 支付者openid
	}
	order.Reason = *resp.TradeStateDesc // 交易状态描述

	if *resp.TradeState == "SUCCESS" {
		order.ResultType = 0
		addDownTimes := 1
		if order.ToolId == 27 { // 27号工具每次付款有两次下载机会
			addDownTimes = 2
		}
		if err := model.AddImgTaskDownTimes(&order, addDownTimes); err != nil { // 增加任务下载次数
			thisLog.Error("更新任务失败", err)
			return
		}
		base.ToolPay(order.ToolId)                                // 记录次数
		base.ToolPayAmount(order.ToolId, *resp.Amount.PayerTotal) // 记录金额
	} else if *resp.TradeState == "REFUND" { // 转入退款
		order.ResultType = 4
	} else if *resp.TradeState == "NOTPAY" { // 未支付
		order.ResultType = 5
	} else if *resp.TradeState == "CLOSED" { // 已关闭
		order.ResultType = 6
	} else if *resp.TradeState == "PAYERROR" { // 支付失败
		order.ResultType = 99
	} else { // 未知状态
		order.ResultType = 99
	}

	if err := order.Update(); err != nil { // 更新订单
		thisLog.Error("更新订单失败", err)
		return
	}
}

// 发起退款
func WechatRefund(order *model.Order, reason string) error {
	thisLog := g.Log{RequestUrl: "WechatRefund"}

	order.ResultType = 4 // 标记为转入退款状态
	order.Reason = reason
	if err := order.Update(); err != nil { // 更新订单
		return err
	}

	_, _, refundErr := refundSvc.Create(context.Background(),
		refunddomestic.CreateRequest{
			OutTradeNo:  core.String(order.Id),
			OutRefundNo: core.String(order.Id),
			Reason:      core.String(reason),
			NotifyUrl:   core.String(fmt.Sprintf("%s/tools/order/wechatRefundNotify", g.VP.GetString("URL"))),
			Amount: &refunddomestic.AmountReq{
				Currency: core.String("CNY"),
				Refund:   &order.Amount,
				Total:    &order.Amount,
			},
		},
	)
	if refundErr != nil {
		thisLog.Error("微信退款失败", refundErr)
		order.Reason = "微信退款失败"
		order.ResultType = 9                   // 标记为退款失败状态
		if err := order.Update(); err != nil { // 更新订单
			return err
		}
		return errors.New("微信退款失败,请联系系统管理员处理")
	}

	return nil
}

// 支付回调
func WechatRefundNotify(c *gin.Context) error {
	temp, _ := c.Get("log")
	thisLog, _ := temp.(*g.Log)

	transaction := new(refunddomestic.Refund)
	res, err := notifyHandler.ParseNotifyRequest(context.Background(), c.Request, transaction)
	// 如果验签未通过，或者解密失败
	if err != nil {
		thisLog.Error("退款回调验签失败", err)
		return err
	}
	// 处理通知内容
	order := model.InitOrder(c, 0, -1) // 因为此处是用于查询订单，所以先toolid先填0，等待后续查询结果
	order.Id = *transaction.OutTradeNo
	if err := order.Get(); err != nil {
		return err
	}
	// 转换订单完成时间
	if transaction.SuccessTime == nil { // 订单未成功时，不会有SuccessTime
		order.EndAt = time.Now().UnixMilli()
	} else {
		order.EndAt = transaction.SuccessTime.UnixMilli()
	}
	if res.EventType == "REFUND.SUCCESS" {
		order.ResultType = 8                                                 // 标记为退款成功状态
		base.ToolRefund(order.ToolId)                                        // 记录次数
		base.ToolRefundAmount(order.ToolId, *transaction.Amount.PayerRefund) // 记录金额
	} else if res.EventType == "REFUND.CLOSED" { // 已关闭
		order.ResultType = 10
	} else if res.EventType == "REFUND.ABNORMAL" { // 退款异常
		order.ResultType = 11
	} else { // 未知状态
		order.ResultType = 99
		thisLog.Error("退款回调未知状态" + order.Id)
		fmt.Println("退款回调信息"+order.Id, res)
	}
	if err := order.Update(); err != nil { // 更新订单
		return err
	}
	if order.ResultType != 8 { // 只有退款成功了，才往下走
		return nil
	}
	// 减少下载次数
	return ToolRefundReduceDownTimes(order)
}

// 从数据库查询常量
func GetWechatConstants(thisLog g.Log) bool {
	mchID = model.SelectConstant(g.VP.GetString("WeChat.MchIdName"))
	if mchID == "" {
		thisLog.Error("微信mchID常量未配置")
		return false
	}
	mchCertNum = model.SelectConstant(g.VP.GetString("WeChat.MchCertName"))
	if mchCertNum == "" {
		thisLog.Error("微信mchCert常量未配置")
		return false
	}
	mchAPIv3Key = model.SelectConstant(g.VP.GetString("WeChat.ApiV3KeyName"))
	if mchAPIv3Key == "" {
		thisLog.Error("微信mchAPIv3Key常量未配置")
		return false
	}
	fansongAppId = model.SelectConstant(g.VP.GetString("WeChat.APPIdName"))
	if fansongAppId == "" {
		thisLog.Error("微信APPId常量未配置")
		return false
	}
	var err error
	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err = utils.LoadPrivateKeyWithPath(g.VP.GetString("WeChat.182Key"))
	if err != nil {
		thisLog.Error("加载微信私钥失败", err)
		return false
	}
	return true
}
