package api

import (
	"net/http"
	"src/app/model"
	"src/app/service/payService"
	g "src/global"
	"src/utils/response"
	"time"

	"github.com/gin-gonic/gin"
)

func WechatpayNotify(c *gin.Context) {
	if err := payService.WechatNotify(c); err != nil {
		// 返回错误信息
		c.AbortWithStatus(500)
		return
	}
	// 返回成功信息
	c.AbortWithStatus(200)
}

func AlipayNotify(c *gin.Context) {
	if err := payService.AlipayNotify(c); err != nil {
		// 返回错误信息
		c.JSON(500, gin.H{
			"code":    "Fail",
			"message": err.Error(),
		})
		return
	}
	// 返回成功信息
	c.String(http.StatusOK, "success")
}

func WechatpayRefundNotify(c *gin.Context) {
	if err := payService.WechatRefundNotify(c); err != nil {
		// 返回错误信息
		c.JSON(500, gin.H{
			"code":    "Fail",
			"message": err.Error(),
		})
		return
	}
	// 返回成功信息
	c.AbortWithStatus(200)
}

type orderRes struct {
	OrderId string `json:"orderId"`
	CodeUrl string `json:"codeUrl"`
}

// 该接口是先创建任务再付款的
func WechatPay(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	if task.FindWithCreateTime(86400); task.ErrMsg != "" {
		response.Fail(c, g.CurdSelectFailCode, task.ErrMsg)
		return
	}
	order := model.InitOrder(c, task.ToolId, 1)
	codeUrl, err := payService.WechatNativeCreateOrder(c, order)
	if err != nil {
		response.Fail(c, g.UnknownErrorCode, err)
		return
	}
	or := orderRes{
		OrderId: order.Id,
		CodeUrl: *codeUrl,
	}
	response.Success(c, or)
}

// 该接口是先创建任务再付款的
func AliH5Pay(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	if task.FindWithCreateTime(86400); task.ErrMsg != "" {
		response.Fail(c, g.CurdSelectFailCode, task.ErrMsg)
		return
	}
	order := model.InitOrder(c, task.ToolId, 2)
	codeUrl, err := payService.AliH5CreateOrder(c, order)
	if err != nil {
		response.Fail(c, g.UnknownErrorCode, err)
		return
	}
	or := orderRes{
		OrderId: order.Id,
		CodeUrl: *codeUrl,
	}
	response.Success(c, or)
}

// 该接口是先付款再创建任务的
func WechatPayFirst(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	if task.ToolId != 27 { // 当前只允许工具27
		response.Fail(c, g.ValidatorParamsCheckFailCode, "该接口不支持")
		return
	}
	task.CreatedAt = time.Now().UnixMilli()
	if task.Add(); task.ErrMsg != "" {
		response.Fail(c, g.CurdSelectFailCode, task.ErrMsg)
		return
	}
	order := model.InitOrder(c, task.ToolId, 1)
	codeUrl, err := payService.WechatNativeCreateOrder(c, order)
	if err != nil {
		response.Fail(c, g.UnknownErrorCode, err)
		return
	}
	or := orderRes{
		OrderId: order.Id,
		CodeUrl: *codeUrl,
	}
	response.Success(c, or)
}

// 该接口是先付款再创建任务的
func AliPayFirst(c *gin.Context) {
	task := model.InitImgTask(c)
	if task.ErrMsg != "" {
		response.Fail(c, g.ValidatorParamsCheckFailCode, task.ErrMsg)
		return
	}
	if task.ToolId != 27 { // 当前只允许工具27
		response.Fail(c, g.ValidatorParamsCheckFailCode, "该接口不支持")
		return
	}
	task.CreatedAt = time.Now().UnixMilli()
	if task.Add(); task.ErrMsg != "" {
		response.Fail(c, g.CurdSelectFailCode, task.ErrMsg)
		return
	}
	order := model.InitOrder(c, task.ToolId, 2)
	codeUrl, err := payService.AliH5CreateOrder(c, order)
	if err != nil {
		response.Fail(c, g.UnknownErrorCode, err)
		return
	}
	or := orderRes{
		OrderId: order.Id,
		CodeUrl: *codeUrl,
	}
	response.Success(c, or)
}

type payRes struct {
	TaskId  string `json:"taskId"`
	Status  uint8  `json:"status"`
	Message string `json:"message"`
}

// 查询订单状态
func PayStatus(c *gin.Context) {
	orderId := c.Param("orderId")
	order := model.InitOrder(c, 0, -1) // todo: 这里应该根据订单号查询订单
	order.Id = orderId
	if err := order.Get(); err != nil {
		response.Fail(c, g.CurdSelectFailCode, err.Error())
		return
	}
	or := payRes{TaskId: order.TaskId, Status: order.ResultType, Message: order.Reason}
	response.Success(c, or)
}
