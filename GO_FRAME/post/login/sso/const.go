// Package sso 是单点登录示例的公共常量包，sso 认证服务与各业务系统（app1等）共享同一份配置
package sso

const (
	// SSO_URL sso认证服务监听地址
	SSO_URL = "127.0.0.1:5678"
	// APP1_URL app1业务系统监听地址
	APP1_URL = "127.0.0.1:5679"

	// SSO_TOKEN_QUERY_NAME sso_token在URL query参数中的名字
	SSO_TOKEN_QUERY_NAME = "sso_token"
	// SSO_TOKEN_COOKIE_NAME sso_token在cookie中的名字
	SSO_TOKEN_COOKIE_NAME = "sso_token"
	// SSO_TOKEN_LIFE sso_token有效期（秒），即JWT过期时间
	SSO_TOKEN_LIFE = 3600

	// APP1_TOKEN_COOKIE_NAME app1本地会话cookie的名字
	APP1_TOKEN_COOKIE_NAME = "app1_token"
	// APP_TOKEN_LIFE app1本地会话有效期（秒）
	APP_TOKEN_LIFE = 86400

	// KEY_OF_UID 用户信息中用户ID的key
	KEY_OF_UID = "user_id"
	// KEY_OF_NAME 用户信息中用户名的key
	KEY_OF_NAME = "user_name"
)
