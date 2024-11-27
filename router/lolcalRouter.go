package router

import (
	"src/app/api"
	"src/app/proxy"
	g "src/global"
	"src/middleware"
	"syscall"

	"github.com/gin-gonic/gin"
)

// 主要是添加一些只允许本地访问的路由
func addLocalR(localR *gin.RouterGroup) {
	localR.Use(middleware.LocalhostHandler())
	// 用于手工停止服务
	localR.GET("/exitService", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "退出服务"})
		ctx.Abort()
		g.ExitSignal <- syscall.SIGTERM
	})
	// 查询ip管理信息的端口
	// routerIP.GET("/get_all", proxy.GetAll)
	// routerIP.GET("/get_permanent_ban", proxy.GetPermanentBan)
	localR.GET("/ip/get_len", proxy.GetLen)
	localR.GET("/ip/get_ban", proxy.GetPermanentBan)
	localR.GET("/ip/get_size", proxy.GetSizeOf)
	localR.GET("/ip/add_ban", proxy.AddBanTime)
	localR.GET("/ip/delete", proxy.DeleteIP) // 删除ip
	localR.GET("/static/update", g.UpdateStaticFile) // 更新静态资源
	localR.GET("/tools/update",api.UpdateToolsMsg) // 更新工具信息
}
