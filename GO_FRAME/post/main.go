package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	database "github.com/teslaXvip/go/go_frame/post/database/gorm"
	handler "github.com/teslaXvip/go/go_frame/post/handler/gin"
	"github.com/teslaXvip/go/go_frame/post/util"
)

func Init() {
	util.InitSlog("./log/post.log")
	database.ConnectPostDB("./post/conf", "db", util.YAML, "./log")
}

func main() {
	Init()

	engine := gin.Default()

	// 静态资源路由
	engine.Static("/js", "post/views/js") // 在url访问目录/js，对应文件系统post/views/js目录
	engine.Static("/css", "post/views/css")
	engine.StaticFile("/favicon.ico", "post/views/img/dqq.png") // url访问/favicon.ico，映射到指定图片文件

	// 加载html模板
	engine.LoadHTMLGlob("post/views/html/*")

	// GET页面路由：访问页面
	engine.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(200, "login.html", nil)
	})
	engine.GET("/regist", func(ctx *gin.Context) {
		ctx.HTML(200, "user_regist.html", nil)
	})
	engine.GET("/modify_pass", func(ctx *gin.Context) {
		ctx.HTML(200, "update_pass.html", nil)
	})

	// POST接口路由：表单提交处理
	engine.POST("/login/submit", handler.Login)
	engine.POST("/modify_pass/submit", handler.Auth, handler.UpdatePassword)
	engine.GET("/logout", handler.Logout)

	group := engine.Group("/news")
	group.GET("", handler.NewsList)
	group.GET("/issue", func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "news_issue.html", nil) })
	group.POST("/issue/submit", handler.Auth, handler.PostNews)
	group.GET("/belong", handler.NewsBelong)
	group.GET("/:id", handler.GetNewsById)
	group.GET("/delete/:id", handler.Auth, handler.DeleteNews)
	group.POST("/update", handler.Auth, handler.UpdateNews)

	engine.GET("", func(ctx *gin.Context) { ctx.Redirect(http.StatusMovedPermanently, "news") }) //新闻列表页是默认的首页

	engine.Run("localhost:5678")
}
