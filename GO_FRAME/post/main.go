package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	database "github.com/teslaXvip/go/go_frame/post/database/gorm"
	handler "github.com/teslaXvip/go/go_frame/post/handler/gin"
	"github.com/teslaXvip/go/go_frame/post/util"
)

func Init() *cron.Cron {
	util.InitSlog("./log/post.log")
	database.ConnectPostDB("./post/conf", "db", util.YAML, "./log")

	crontab := cron.New()
	// 定时表达式：30-50/5 * 1,4,8,10-20 1-6 *
	// 任务：调用 database.PingPostDB
	_, err := crontab.AddFunc("*/1 * * * *", database.PingPostDB)
	if err != nil {
		panic(err)
	}

	// 启动定时任务
	crontab.Start()
	return crontab
}

func ListenTermSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	slog.Info("receive signal" + sig.String() + ", going to exit")
	database.ClosePostDb()
	os.Exit(0)
}

func main() {
	crontab := Init()
	defer crontab.Stop()

	go ListenTermSignal()

	engine := gin.Default()

	// 静态资源路由
	engine.Static("/js", "post/views/js") // 在url访问目录/js，对应文件系统post/views/js目录
	engine.Static("/css", "post/views/css")
	engine.StaticFile("/favicon.ico", "post/views/img/dqq.png") // url访问/favicon.ico，映射到指定图片文件

	// 加载html模板
	engine.LoadHTMLGlob("post/views/html/*")

	engine.Use(handler.Metric()) // 全局中间件, 上报每一个接口的耗时和调用次数

	engine.GET("/metrics", func(ctx *gin.Context) { //Prometheus要来访问这个接口
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})

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
	engine.POST("/regist/submit", handler.RegistUser)
	engine.POST("/login/submit", handler.Login)
	engine.POST("/modify_pass/submit", handler.Auth, handler.UpdatePassword)
	engine.GET("/logout", handler.Logout)

	group := engine.Group("/news")
	group.GET("", handler.NewsList)
	group.GET("/issue", func(ctx *gin.Context) { ctx.HTML(http.StatusOK, "news_issue.html", nil) })
	group.POST("/issue/submit", handler.Auth, handler.PostNews)
	group.GET("/belong", handler.NewsBelong)
	group.GET("/edit/:id", handler.Auth, handler.EditNewsPage)
	group.GET("/:id", handler.GetNewsById)
	group.GET("/delete/:id", handler.Auth, handler.DeleteNews)
	group.POST("/update", handler.Auth, handler.UpdateNews)

	engine.GET("", func(ctx *gin.Context) { ctx.Redirect(http.StatusMovedPermanently, "news") }) //新闻列表页是默认的首页

	engine.Run("localhost:5678")
}
