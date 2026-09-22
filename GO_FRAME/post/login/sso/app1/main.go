package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	sso "github.com/teslaXvip/go/go_frame/post/login/sso"
)

// app1本地会话redis的key前缀，与其他app区分
const (
	// Session相对于JWT的好处: Server端发现有风险时,随时可以撤回Session,让用户重新登录
	SESSION_KEY_PREFIX = "app1_session_"
)

// Redis 客户端，包级初始化，地址可按需调整
var rdb = redis.NewClient(&redis.Options{
	Addr: "127.0.0.1:6379",
})

// newSessionID 用 crypto/rand 生成16字节随机数，hex编码为32位，不引入第三方依赖
func newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// checkToken 向sso服务校验sso_token，拿到用户信息
func checkToken(token string) (userName, userId string, valid bool) {
	resp, err := http.Get("http://" + sso.SSO_URL + "/identify?" + sso.SSO_TOKEN_QUERY_NAME + "=" + token)
	if err != nil {
		slog.Error("请求sso校验token失败", "error", err)
		return "", "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("sso校验token返回非200", "status", resp.StatusCode)
		return "", "", false
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("读取sso响应失败", "error", err)
		return "", "", false
	}
	var mp map[string]string
	if err := json.Unmarshal(bs, &mp); err != nil {
		slog.Error("解析sso响应失败", "error", err)
		return "", "", false
	}
	return mp[sso.KEY_OF_NAME], mp[sso.KEY_OF_UID], true
}

func Home(ctx *gin.Context) {
	// 先检查cookie有没有携带app_token。
	if cookie, err := ctx.Request.Cookie(sso.APP1_TOKEN_COOKIE_NAME); err == nil {
		sessionID := cookie.Value
		result := rdb.Get(context.Background(), SESSION_KEY_PREFIX+sessionID)
		if result.Err() == nil {
			var mp map[string]string
			if err := json.Unmarshal([]byte(result.Val()), &mp); err == nil {
				userId := mp[sso.KEY_OF_UID]
				userName := mp[sso.KEY_OF_NAME]
				slog.Info("根据app token,身份验证成功")
				ctx.String(200, "这里是app1, 欢迎 "+userName+"["+userId+"]")
				return
			}
		} else if result.Err() != redis.Nil {
			// 会话不存在是正常情况，只记录真正的redis错误
			slog.Error("读取app1会话失败", "sessionID", sessionID, "error", result.Err())
		}
	}

	// 从cookie或者query参数时获得sso_token
	var token string
	if cookie, err := ctx.Request.Cookie(sso.SSO_TOKEN_COOKIE_NAME); err == nil {
		token = cookie.Value
	} else {
		token = ctx.Query(sso.SSO_TOKEN_QUERY_NAME)
	}
	//如果client携带了sso_token
	if len(token) > 0 {
		userName, userId, valid := checkToken(token)
		if valid { //sso_token合法
			slog.Info("根据sso token,身份验证成功")
			ctx.SetCookie(
				sso.SSO_TOKEN_COOKIE_NAME,
				token,
				sso.SSO_TOKEN_LIFE,
				"/",
				"localhost",
				false,
				true,
			)
			sessionID, err := newSessionID() //生成一个随机的字符串
			if err != nil {
				slog.Error("生成sessionID失败", "error", err)
				ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
				return
			}
			ctx.SetCookie(
				sso.APP1_TOKEN_COOKIE_NAME,
				sessionID,
				sso.APP_TOKEN_LIFE,
				"/",
				"localhost",
				false,
				true,
			)
			info, err := json.Marshal(map[string]string{sso.KEY_OF_UID: userId, sso.KEY_OF_NAME: userName})
			if err != nil {
				slog.Error("序列化用户信息失败", "error", err)
				ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
				return
			}
			//key是sessionID, value是sso的用户身份信息
			if err := rdb.Set(context.Background(), SESSION_KEY_PREFIX+sessionID, string(info), sso.APP_TOKEN_LIFE*time.Second).Err(); err != nil {
				slog.Error("写入app1会话到redis失败", "error", err)
				ctx.String(http.StatusInternalServerError, "登录失败，请稍后重试")
				return
			}
			ctx.String(200, "这里是app1, 欢迎 "+userName+"["+userId+"]")
			return
		} else { //sso_token非法, 可能是过期了
			ctx.SetCookie(
				sso.SSO_TOKEN_COOKIE_NAME,
				"",
				-1, //删除cookie
				"/",
				"localhost",
				false,
				true,
			)
			url := "http://" + sso.SSO_URL + "/login?service=" + sso.APP1_URL + "/home"
			ctx.Redirect(http.StatusFound, url)
			return
		}
	} else {
		url := "http://" + sso.SSO_URL + "/login?service=" + sso.APP1_URL + "/home"
		ctx.Redirect(http.StatusFound, url)
		return
	}
}

func main() {
	engine := gin.Default()
	engine.GET("/home", Home)
	if err := engine.Run(sso.APP1_URL); err != nil {
		panic(err)
	}
}
