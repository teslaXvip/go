package util

import (
	"testing"
	"time"
)

func TestGenJWT(t *testing.T) {
	header := JwtHeader{
		Algo: "HS256",
		Type: "JWT",
	}
	payload := JwtPayload{
		ID:         "uuid-001",
		Issue:      "my-server",
		Audience:   "my-client",
		Subject:    "user-login",
		IssueAt:    time.Now().Unix(),
		Expiration: time.Now().Add(time.Hour * 2).Unix(), //2小时过期
		UserDefined: map[string]any{
			"userId": 10001,
			"name":   "zhangsan",
		},
	}
	secret := "my-secret-key-123456"
	token, err := GenJWT(header, payload, secret)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("生成token:", token)

	// 校验token
	h, p, err := VerifyJwt(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("header:%+v", h)
	t.Logf("payload:%+v", p)
}
