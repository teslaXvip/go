package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	REFRESH_KEY_PREFIX  = "session_"
	REFRESH_TOKEN_LIFE  = 60 // refresh token有效期，单位秒
	REFRESH_COOKIE_NAME = "refresh"
	ACCESS_COOKIE_NAME  = "access"
	SECRET              = "f4398" // jwt签名密钥
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

// newRefreshToken 用 crypto/rand 生成16字节随机数，hex编码为32位，不引入第三方依赖
func newRefreshToken() (string, error) {
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
	// 查询数据库用户
	user2 := database.GetUserByName(form.Name)
	if user2 == nil {
		ctx.String(http.StatusBadRequest, "用户名不存在")
		return
	}
	if user2.PassWord != form.PassWord {
		ctx.String(http.StatusBadRequest, "密码错误")
		return
	}

	// 登录成功：生成refreshToken（随机字符串）
	refreshToken, err := newRefreshToken()
	if err != nil {
		slog.Error("生成refreshToken失败", "error", err)
		ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
		return
	}
	// 设置refresh cookie
	ctx.SetCookie(REFRESH_COOKIE_NAME, refreshToken, REFRESH_TOKEN_LIFE, "/", "localhost", false, true)

	// 生成access JWT
	header := util.DefaultHeader
	payload := util.JwtPayload{
		Issue:       "dual_token",
		IssueAt:     time.Now().Unix(),
		Expiration:  0, // JWT本身永不过期！靠cookie会话机制控制
		UserDefined: map[string]any{"user_id": strconv.Itoa(user2.Id), "user_name": user2.Name},
	}
	accessToken, err := util.GenJWT(header, payload, SECRET)
	if err != nil {
		slog.Error("生成access token失败", "error", err)
		ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
		return
	}
	// access存入cookie：0=会话cookie，关浏览器失效
	ctx.SetCookie(ACCESS_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
	// redis存储refreshToken -> accessToken，过期时间60秒
	if err := rdb.Set(context.Background(), REFRESH_KEY_PREFIX+refreshToken, accessToken, REFRESH_TOKEN_LIFE*time.Second).Err(); err != nil {
		slog.Error("写入refresh映射到redis失败", "error", err)
		ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
		return
	}
	// 重定向到需要登录的页面
	ctx.Redirect(http.StatusFound, "/page1")
}

func AuthMiddleWare(ctx *gin.Context) {
	// 第一步：尝试读取access cookie，校验JWT
	if cookie, err := ctx.Request.Cookie(ACCESS_COOKIE_NAME); err == nil {
		accessToken := cookie.Value
		_, payload, err := util.VerifyJwt(accessToken, SECRET)
		if err == nil {
			slog.Info("直接根据access token拿到了用户的身份信息")
			// 用户信息放入gin的上下文，后续接口可以读取
			ctx.Set("user_name", payload.UserDefined["user_name"])
			ctx.Set("user_id", payload.UserDefined["user_id"])
			ctx.Next() // 放行！执行后面的业务函数
			return
		}
	} else {
		slog.Info("cookie里没有access token")
	}

	// 第二步：access校验失败，尝试用refresh token刷新
	if cookie, err := ctx.Request.Cookie(REFRESH_COOKIE_NAME); err == nil {
		refreshToken := cookie.Value
		// 去redis查询refresh对应的accessToken
		result := rdb.Get(context.Background(), REFRESH_KEY_PREFIX+refreshToken)
		if result.Err() == nil {
			slog.Info("根据cookie里的refresh token重新获得了access token")
			accessToken := result.Val()
			_, payload, err := util.VerifyJwt(accessToken, SECRET)
			if err == nil {
				// 把access重新写回浏览器cookie
				ctx.SetCookie(ACCESS_COOKIE_NAME, accessToken, 0, "/", "localhost", false, true)
				slog.Info("把access token种到了浏览器的cookie里")
				ctx.Set("user_name", payload.UserDefined["user_name"])
				ctx.Set("user_id", payload.UserDefined["user_id"])
				ctx.Next() // 刷新成功放行
				return
			}
		} else if result.Err() != redis.Nil {
			// refresh失效是正常情况，只记录真正的redis错误
			slog.Error("读取refresh映射失败", "error", result.Err())
		}
	}

	// 第三步：access和refresh都失效，跳转到登录页
	ctx.Redirect(http.StatusFound, "/login")
	ctx.Abort() // 终止后续handler执行
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

	router.Static("/js", "post/views/js") //在url是访问目录/js相当于访问文件系统中的views/js目录
	router.Static("/css", "post/views/css")
	router.StaticFile("/favicon.ico", "post/views/img/dqq.png") //在url中访问文件/favicon.ico, 相当于访问文件系统中的views/img/dqq.png文件
	router.LoadHTMLGlob("post/views/html/*")                    //使用这些.html文件时就不需要加路径了

	// 登录页面
	router.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "session_login.html", nil)
	})
	// 登录
	router.POST("/login/submit", Login)
	// 需要登录才能访问的页面1
	router.GET("/page1", AuthMiddleWare, func(ctx *gin.Context) { // 使用身份认证中间件
		// 从ctx里取出用户信息
		uname, _ := ctx.Get("user_name")
		uid, _ := ctx.Get("user_id")
		ctx.String(200, "这是page1, 欢迎 "+uname.(string)+"["+uid.(string)+"]")
	})
	// 需要登录才能访问的页面2
	router.GET("/page2", AuthMiddleWare, func(ctx *gin.Context) { // 使用身份认证中间件
		// 从ctx里取出用户信息
		uname, _ := ctx.Get("user_name")
		uid, _ := ctx.Get("user_id")
		ctx.String(200, "这是page2, 欢迎 "+uname.(string)+"["+uid.(string)+"]")
	})

	if err := router.Run("127.0.0.1:5678"); err != nil {
		panic(err)
	}
}
