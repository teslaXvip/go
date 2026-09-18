package handler

import (
	"net/http"

	"github.com/teslaXvip/go/go_frame/post/util"

	"github.com/gin-gonic/gin"
)

var (
	KeyConfig = util.InitViper("post/conf", "jwt", util.YAML)
)

const (
	COOKIE_NAME  = "uid"  // cookie名称，存放JWT登录凭证
	COOKIE_LIFE  = 604800 // cookie有效期（秒），7天
	UID_IN_TOKEN = "uid"  // JWT payload自定义字段key，存放用户ID
	UID_IN_CTX   = "uid"  // Gin上下文key，存放登录用户ID
)

// GetLoginUid 从ctx的Cookie中取出jwt token，再解析出uid
func GetLoginUid(ctx *gin.Context) int {
	token := ""
	// 遍历所有cookie，找到名字等于COOKIE_NAME的cookie
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name == COOKIE_NAME {
			token = cookie.Value
		}
	}
	// 交给函数解析token得到uid
	return GetUidFromJwt(token)
}

// GetUidFromJwt 校验JWT签名，解析payload，取出uid
func GetUidFromJwt(token string) int {
	// 调用util包的VerifyJwt，校验签名，拿到payload
	_, payload, err := util.VerifyJwt(token, KeyConfig.GetString("secret"))
	if err != nil {
		// token不存在、过期、签名错误，返回0（代表未登录）
		return 0
	}
	// 遍历自定义字段，找到uid
	for k, v := range payload.UserDefined {
		if k == UID_IN_TOKEN {
			// json解析数字默认是float64，强制转int
			return int(v.(float64))
		}
	}
	return 0
}

// Auth 身份认证中间件，Gin HandlerFunc类型
func Auth(ctx *gin.Context) {
	loginUid := GetLoginUid(ctx)
	if loginUid <= 0 {
		// 未登录：重定向到登录页面
		ctx.Redirect(http.StatusTemporaryRedirect, "/login")
		ctx.Abort() // 终止后续handler执行
		return      // Abort不会停止当前函数，必须return
	} else {
		// 校验成功，把uid存入Gin上下文，后续接口可以 ctx.Get(UID_IN_CTX) 获取
		ctx.Set(UID_IN_CTX, loginUid)
	}
}
