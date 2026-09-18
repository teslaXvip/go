package handler

import (
	"net/http"
	"strconv"

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
	ctx.SetCookie("uid", strconv.Itoa(user2.Id), 86400, "/", "localhost", false, true)
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
