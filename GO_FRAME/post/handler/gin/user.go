package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	database "github.com/teslaXvip/go/go_frame/post/database/gorm"
	"github.com/teslaXvip/go/go_frame/post/handler/model"
	"github.com/teslaXvip/go/go_frame/post/util"
)

func RegistUser(ctx *gin.Context) {
	var user model.User
	err := ctx.ShouldBind(&user)
	if err != nil {
		ctx.String(http.StatusBadRequest, util.BindingErrMsg(err))
		return
	}
	_, err = database.RegistUser(user.Name, user.PassWord)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	// 到这 gin会自动返回200
	// ctx.Status(200)
}

func Login(ctx *gin.Context) {
	var user model.User
	err := ctx.ShouldBind(&user)
	if err != nil {
		ctx.String(http.StatusBadRequest, util.BindingErrMsg(err))
		return
	}
	user2 := database.GetUserByName(user.Name)
	if user2 == nil {
		ctx.String(http.StatusBadRequest, "用户名不存在")
		return
	}

	if user2.PassWord == user.PassWord {
		ctx.String(http.StatusBadRequest, "密码错误")
		return
	}

	// 登录成功，构造JWT
	header := util.DefaultHeader // 工具包预先定义好默认JwtHeader: HS256,JWT
	payload := util.JwtPayload{
		Issue:      "news",
		IssueAt:    time.Now().Unix(),                                //签发时间，每次登录时间不一样 → token每次不一样
		Expiration: time.Now().Add(COOKIE_LIFE * time.Second).Unix(), //7天后过期
		UserDefined: map[string]any{
			UID_IN_TOKEN: user2.Id, //自定义字段，存放用户ID
		},
	}

	// 生成JWT Token
	if token, err := util.GenJWT(header, payload, KeyConfig.GetString("secret")); err != nil {
		slog.Error("生成token失败", "error", err)
		ctx.String(http.StatusInternalServerError, "token生成失败")
	} else {
		// Gin 设置Cookie，返回给浏览器 Set-Cookie响应头
		ctx.SetCookie(
			COOKIE_NAME, // cookie名称
			token,       // cookie值：JWT字符串
			COOKIE_LIFE, // maxAge 有效期秒数，>0代表多少秒后过期
			"/",         // path: 整个网站路径都带上这个cookie
			"localhost", // domain：仅localhost域名下携带cookie
			false,       // secure：false http可访问；true仅https
			true,        // HttpOnly: true，禁止JS读取/修改cookie，防XSS
		)
		ctx.String(http.StatusOK, "登录成功")
	}
}

func Logout(ctx *gin.Context) {
	ctx.SetCookie("uid", "", -1, "/", "localhost", false, true)
}

func UpdatePassword(ctx *gin.Context) {
	var req model.ModifyPassRequest
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.String(http.StatusBadRequest, util.BindingErrMsg(err))
		return
	}
	uid := GetUidFromCookie(ctx)

	if uid <= 0 {
		ctx.String(http.StatusForbidden, "请先登陆")
		return
	}

	err = database.UpdatePassword(uid, req.OldPass, req.NewPass)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
}

func GetUidFromCookie(ctx *gin.Context) int {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name == "uid" {
			uid, err := strconv.Atoi(cookie.Value)
			if err == nil {
				return uid
			}
		}
	}
	return 0
}
