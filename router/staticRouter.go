package router

import (
	"src/app/proxy"
	"src/middleware"

	"github.com/gin-gonic/gin"
)

func addStaticRoutes(r *gin.RouterGroup) {

	// 代理js，css等文件，因为数量众多，所以不计入限制的中间件中
	r.GET("/dist/*name", proxy.ProxyStatic)

	r.Use(middleware.LimitHandler())
	// 下面这些资源加入访问次数控制，上面的dist不加，因为访问频率太高
	r.GET("/", proxy.ProxyStatic)
	r.GET("/images/*name", proxy.ProxyStatic)
	r.GET("/home/*name", proxy.ProxyStatic)
	r.GET("/image/*name", proxy.ProxyStatic)
	r.StaticFile("/sitemap.xml", "./sitemap/sitemap.xml")

	// 这几个路由，是防止用户末尾不输入/，从而导致匹配不到上面路由
	r.GET("/image", proxy.ProxyStatic)
	r.GET("/home", proxy.ProxyStatic)
}
