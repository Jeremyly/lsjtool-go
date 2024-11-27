package model

import (
	"context"
	"errors"
	g "src/global"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yitter/idgenerator-go/idgen"
)

type Order struct {
	BaseModel
	Id            string `gorm:"primaryKey"`                         // 订单号，使用雪花算法生成
	ToolId        uint8  `json:"tool_id" db:"tool_id"`               // 工具id
	TaskId        string `json:"task_id" db:"task_id"`               // 系统内部任务id
	Platform      uint8  `json:"platform" db:"platform"`             // 平台类型：1:表示微信；2表示支付宝
	TradeNo       string `json:"trade_no" db:"trade_no"`             // 订单号，外部平台生成的唯一订单号
	Amount        int64  `json:"amount" db:"amount"`                 // 订单金额
	CreatedAt     int64  `gorm:"autoCreateTime:false"`               // 订单创建时间，记录的是平台返回支付链接的时间，如果创建订单时返回错误，则该时间为空
	ResultType    uint8  `json:"result_type" db:"result_type"`       // 0：支付成功 1：支付中 2：订单创建失败 3：取消支付 4：转入退款 5：未支付 6：已关闭 7：已撤销 8：已退款 999：支付失败（带其他具体原因）
	Reason        string `json:"reason" db:"reason"`                 // 失败原因
	EndAt         int64  `json:"end_at" db:"end_at"`                 // 订单结束时间。可以是创建订单失败时响应的时间，也可以是发起支付后有结果的收到消息的时间，无论该结果是否是成功支付
	TradeType     string `json:"trade_type" db:"trade_type"`         // 建议类型。 微信：JSAPI：公众号支付 NATIVE：扫码支付 App：App支付 MICROPAY：付款码支付 MWEB：H5支付 FACEPAY：刷脸支付
	Payer         string `json:"-" db:"payer"`                       // 支付人标识
	RefundChannel string `json:"refund_channel" db:"refund_channel"` // 退款渠道
}

func InitOrder(c *gin.Context, toolId uint8, platform int, ctx ...context.Context) *Order {
	taskId := c.Request.Header.Get(g.TaskIdKey)

	return &Order{Id: strconv.FormatInt(idgen.NextId(), 10),
		ToolId:    toolId,
		TaskId:    taskId,
		Platform: uint8(platform),
		Amount:    int64(ToolsMap[toolId].DiscountPrice * 100),
		BaseModel: *InitBaseModel(c).GetCtx(c, ctx...).GetLog(c).GetDB(c, ctx...),
	}
}

func (o *Order) Insert() {
	if err := o.DB.Create(o).Error; err != nil {
		o.Log.Error("插入订单失败", err)
	}
}

func (o *Order) Get() error {
	if o.Id == "" {
		return errors.New("订单ID不能为空！")
	}
	if err := o.DB.First(o).Error; err != nil {
		o.Log.Error("查询订单失败", err)
		return err
	}
	if o.ToolId == 0 {
		return errors.New("未找到该订单！")
	}
	return nil
}

func (o *Order) Update() error {
	if err := o.DB.Save(o).Error; err != nil {
		o.Log.Error("更新订单失败", err)
		return err
	}
	return nil
}

// 检查是否付款
func GetOrderByTaskId(task *ImgTask) (*Order, error) {
	order := Order{TaskId: task.Id}
	if err := task.DB.Table("order").Where("task_id = ?", task.Id).First(&order).Error; err != nil {
		task.Log.Error("查询订单失败", err)
		return nil, errors.New("查询订单失败")
	}
	if order.ToolId == 0 { // 需要先判断是否付款，否则无此订单ResultType默认值也是0
		return nil, errors.New("未找到该订单！")
	}
	order.Log = task.Log
	order.DB = task.DB
	return &order, nil
}
