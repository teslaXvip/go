package io

import (
	"fmt"
	"time"
)

// channel类型的数据不支持序列化

var MyDateFormat = "2006-01-02"

/*
参考时间 01/02 03:04:05PM '06 -0700 可以这样记：
 1月2日 3点4分5秒 下午 06年 时区
*/

// MyDate 基于time.Time自定义类型
type MyDate time.Time

// MarshalJSON 自定义序列化（值接收者）
func (d MyDate) MarshalJSON() ([]byte, error) {
	// \" 转义双引号
	s := fmt.Sprintf("\"%s\"", time.Time(d).Format(MyDateFormat))
	return []byte(s), nil
}

// UnmarshalJSON 自定义反序列化（指针接收者，必须指针才能修改原值）
func (d *MyDate) UnmarshalJSON(bs []byte) (err error) {
	// bs 自带双引号，所以模板格式前后也要带上引号
	now, err := time.ParseInLocation(`"`+MyDateFormat+`"`, string(bs), time.Local)
	//layout:  '"2006-01-02"' （注意外层是反引号字符串）
	if err != nil {
		return err
	}
	*d = MyDate(now)
	return nil
}

// 测试结构体
type User struct {
	Name  string `json:"name"`
	Birth MyDate `json:"birth"`
}
