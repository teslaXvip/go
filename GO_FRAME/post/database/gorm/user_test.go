package database_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"

	database "github.com/teslaXvip/go/go_frame/post/database/gorm"
	"github.com/teslaXvip/go/go_frame/post/util"
)

// init：包初始化函数，执行go test的时候会优先执行init()
func init() {
	// 初始化日志，注意测试文件的相对路径！
	util.InitSlog("../../../log/post.log")
	// 连接Post数据库：配置目录、配置key、配置文件类型、日志目录
	database.ConnectPostDB("../../conf", "db", util.YAML, "../../../log")
}

// hash：对密码做MD5哈希，返回32位16进制字符串
func hash(pass string) string {
	hasher := md5.New()
	hasher.Write([]byte(pass))
	digest := hasher.Sum(nil)
	// md5输出128bit二进制，hex编码之后得到32位字符串
	return hex.EncodeToString(digest)
}

func TestRegistUser(t *testing.T) {
	// 测试1：正常注册新用户
	uid, err := database.RegistUser("dqq", hash("123456"))
	if err != nil {
		t.Fatal(err) // 出错直接终止测试
	} else {
		fmt.Printf("注册成功，uid=%d\n", uid)
	}

	// 测试2：重复注册同一个用户名，预期报错
	uid, err = database.RegistUser("dqq", hash("123456"))
	if err == nil {
		// 没有报错，说明不符合预期，测试失败
		fmt.Println("重复注册成功! ")
		t.Fail()
	} else {
		fmt.Printf("注册失败: %s\n", err)
	}
}
