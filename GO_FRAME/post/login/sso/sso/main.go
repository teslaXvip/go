package main

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	database "github.com/teslaXvip/go/go_frame/post/database/gorm"
	sso "github.com/teslaXvip/go/go_frame/post/login/sso"
	"github.com/teslaXvip/go/go_frame/post/util"
)

const SECRET = "8j9huumu"

// loginForm 表单绑定，字段名与主项目 handler/model 保持一致（name/pass，密码为 md5 32位）
type loginForm struct {
	Name     string `form:"name" binding:"required,gte=2"`
	PassWord string `form:"pass" binding:"required,len=32"`
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

	//登录成功, 生成JWT token
	header := util.DefaultHeader
	payload := util.JwtPayload{
		Issue:      "sso",
		IssueAt:    time.Now().Unix(),
		Expiration: time.Now().Add(sso.SSO_TOKEN_LIFE * time.Second).Unix(),
		UserDefined: map[string]any{
			sso.KEY_OF_UID:  strconv.Itoa(user2.Id),
			sso.KEY_OF_NAME: user2.Name,
		},
	}
	token, err := util.GenJWT(header, payload, SECRET)
	if err != nil {
		slog.Error("生成token失败", "error", err)
		ctx.String(http.StatusInternalServerError, "生成token失败")
		return
	}
	// 登录成功，把token返回给页面，由页面携带sso_token跳回业务系统
	ctx.String(http.StatusOK, token)
}

// Identify 验证一个token是否合法。如果不合法, 返回code=403; 如果合法返回json,包含uid和name
func Identify(ctx *gin.Context) {
	token := ctx.Query(sso.SSO_TOKEN_QUERY_NAME)
	_, payload, err := util.VerifyJwt(token, SECRET)
	if err != nil {
		ctx.String(http.StatusForbidden, "token验证失败")
		return
	}
	var uid string
	var name string
	if v, exists := payload.UserDefined[sso.KEY_OF_UID]; exists {
		uid = v.(string)
	}
	if v, exists := payload.UserDefined[sso.KEY_OF_NAME]; exists {
		name = v.(string)
	}
	if uid == "" || name == "" {
		ctx.String(http.StatusForbidden, "token验证失败")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{sso.KEY_OF_UID: uid, sso.KEY_OF_NAME: name})
}

func main() {
	database.ConnectPostDB("./post/conf", "db", util.YAML, "./log")

	engine := gin.Default()
	engine.Static("/js", "post/views/js") //在url是访问目录/js相当于访问文件系统中的views/js目录
	engine.Static("/css", "post/views/css")
	engine.StaticFile("/favicon.ico", "post/views/img/dqq.png") //在url中访问文件/favicon.ico, 相当于访问文件系统中的views/img/dqq.png文件
	engine.LoadHTMLGlob("post/views/html/*")                    //使用这些.html文件时就不需要加路径了

	// 登录页面
	engine.GET("/login", func(ctx *gin.Context) {
		service := ctx.Query("service")
		ctx.HTML(http.StatusOK, "oss_login.html", gin.H{"service": service})
	})
	// 登录
	engine.POST("/login/submit", Login)
	// 验证一个token是否合法
	engine.GET("/identify", Identify)
	if err := engine.Run(sso.SSO_URL); err != nil {
		panic(err)
	}
}

// go run ./post/login/sso/sso
