package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	database "github.com/teslaXvip/go/go_frame/post/database/gorm"
	"github.com/teslaXvip/go/go_frame/post/util"
)

const (
	SESSION_KEY_PREFIX = "session_"
	SESSION_LIFE       = 86400       // 24小时，单位秒
	COOKIE_NAME        = "sesion_id" // 注意：这里拼写错误！session写成sesion
)

// Redis 客户端，包级初始化，地址可按需调整
var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
})

// loginForm 表单绑定，字段名与主项目 handler/model 保持一致（name/pass，密码为 md5 32位）
type loginForm struct {
	Name     string `form:"name" binding:"required,gte=2"`
	PassWord string `form:"pass" binding:"required,len=32"`
}

type userInfo struct {
	Name string
	Id   int
}

// newSessionID 用 crypto/rand 生成16字节随机数，hex编码为32位，不引入第三方依赖
func newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func Login(ctx *gin.Context) {
	var form loginForm
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.String(http.StatusBadRequest, util.BindingErrMsg(err))
		return
	}
	user2 := database.GetUserByName(form.Name)
	if user2 == nil {
		ctx.String(http.StatusBadRequest, "用户名不存在")
		return
	}
	if user2.PassWord != form.PassWord {
		ctx.String(http.StatusBadRequest, "密码错误")
		return
	}

	// 登录成功，生成sessionId
	sessionID, err := newSessionID()
	if err != nil {
		slog.Error("生成sessionID失败", "error", err)
		ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
		return
	}
	// 设置Cookie返回浏览器
	ctx.SetCookie(
		COOKIE_NAME,
		sessionID,
		SESSION_LIFE,
		"/",
		"localhost",
		false,
		true, // HttpOnly=true，JS不能读取cookie，防XSS，推荐
	)

	// 用户信息序列化存入Redis
	info, err := json.Marshal(userInfo{Name: user2.Name, Id: user2.Id})
	if err != nil {
		slog.Error("序列化用户信息失败", "name", user2.Name, "error", err)
		ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
		return
	}
	if err := rdb.Set(
		context.Background(),
		SESSION_KEY_PREFIX+sessionID,
		string(info),
		SESSION_LIFE*time.Second,
	).Err(); err != nil {
		slog.Error("写入session到redis失败", "sessionID", sessionID, "error", err)
		ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
		return
	}
	// 跳转到需要登录的页面
	ctx.Redirect(http.StatusFound, "/page1")
}

func AuthMiddleWare(ctx *gin.Context) {
	if cookie, err := ctx.Request.Cookie(COOKIE_NAME); err == nil {
		sessionID := cookie.Value
		result := rdb.Get(context.Background(), SESSION_KEY_PREFIX+sessionID)
		if result.Err() == nil {
			info := result.Val()
			var user userInfo
			if err := json.Unmarshal([]byte(info), &user); err == nil {
				ctx.Set("user_name", user.Name)
				ctx.Set("user_id", strconv.Itoa(user.Id))
				ctx.Next() // 放行！执行后面的业务函数
				return
			}
		} else if result.Err() != redis.Nil {
			// session不存在或已过期是正常情况，只记录真正的redis错误
			slog.Error("读取session失败", "sessionID", sessionID, "error", result.Err())
		}
	}
	// 认证失败跳转登录
	ctx.Redirect(http.StatusFound, "/login")
}

func main() {
	database.ConnectPostDB("./post/conf", "db", util.YAML, "./log")
	// 检查Redis连通性，失败只记录日志，具体错误在登录时再反馈给用户
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("连接Redis失败", "error", err)
	} else {
		slog.Info("连接Redis成功")
	}
	router := gin.Default()

	// 静态资源
	router.Static("/js", "post/views/js")
	router.Static("/css", "post/views/css")
	router.StaticFile("/favicon.ico", "post/views/img/dqq.png")
	router.LoadHTMLGlob("post/views/html/*")

	// 登录页面 GET
	router.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "session_login.html", nil)
	})
	// 登录提交接口 POST
	router.POST("/login/submit", Login)

	// 需要登录的路由，中间件在前
	router.GET("/page1", AuthMiddleWare, func(ctx *gin.Context) {
		// 从ctx取出登录时存入的用户信息
		uname, _ := ctx.Get("user_name")
		uid, _ := ctx.Get("user_id")
		ctx.String(200, "这是page1，欢迎 "+uname.(string)+" ["+uid.(string)+"]")
	})
	router.GET("/page2", AuthMiddleWare, func(ctx *gin.Context) {
		uname, _ := ctx.Get("user_name")
		uid, _ := ctx.Get("user_id")
		ctx.String(200, "这是page2，欢迎 "+uname.(string)+" ["+uid.(string)+"]")
	})

	if err := router.Run("127.0.0.1:5678"); err != nil {
		panic(err)
	}
}
