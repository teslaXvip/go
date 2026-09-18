package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JwtHeader jwt头部
type JwtHeader struct {
	Algo string `json:"alg"` //哈希算法,默认为HMAC SHA256(写为 HS256)
	Type string `json:"typ"` //令牌(token)类型,统一写为JWT
}

// DefaultHeader 默认JWT头部：HS256,JWT
var DefaultHeader = JwtHeader{
	Algo: "HS256",
	Type: "JWT",
}

// JwtPayload jwt载荷，存放声明
type JwtPayload struct {
	ID          string         `json:"jti"` //JWT ID用于标识该JWT
	Issue       string         `json:"iss"` //发行人。比如微信
	Audience    string         `json:"aud"` //受众人。比如王者荣耀
	Subject     string         `json:"sub"` //主题
	IssueAt     int64          `json:"iat"` //发布时间,精确到秒
	NotBefore   int64          `json:"nbf"` //在此之前不可用,精确到秒
	Expiration  int64          `json:"exp"` //到期时间,精确到秒
	UserDefined map[string]any `json:"ud"`  //用户自定义的其他字段
}

// GenJWT 生成JWT token
func GenJWT(header JwtHeader, payload JwtPayload, secret string) (string, error) {
	var part1, part2, signature string

	// 1.header转json -> RawURLBase64编码
	bs1, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	part1 = base64.RawURLEncoding.EncodeToString(bs1)

	// 2.payload转json -> RawURLBase64编码
	bs2, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	part2 = base64.RawURLEncoding.EncodeToString(bs2)

	// 3.签名：HMAC-SHA256(part1.part2, secret密钥)，结果再RawURLBase64
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(part1 + "." + part2))
	signature = base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	// 拼接三部分
	return part1 + "." + part2 + "." + signature, nil
}

// VerifyJwt 校验JWT token
func VerifyJwt(token string, secret string) (*JwtHeader, *JwtPayload, error) {
	// 1.按 . 分割token，必须是3段
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("token是%d部分", len(parts))
	}

	// 2.【核心校验】重新计算签名，和传入的signature对比
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(parts[0] + "." + parts[1]))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	if signature != parts[2] {
		return nil, nil, fmt.Errorf("验证失败")
	}

	// 3.base64解码header和payload
	var part1, part2 []byte
	var err error
	part1, err = base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("header Base64反解失败")
	}
	part2, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("payload Base64反解失败")
	}

	// 4.json反序列化到结构体
	var header JwtHeader
	var payload JwtPayload
	err = json.Unmarshal(part1, &header)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(part2, &payload)
	if err != nil {
		return nil, nil, err
	}

	return &header, &payload, nil
}
