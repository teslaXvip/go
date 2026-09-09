package io_test

import (
	"fmt"
	"go_frame/io"
	"testing"
)

func TestViper(t *testing.T) {
	dbViper := io.InitViper("../conf", "mysql", "yaml")
	dbViper.WatchConfig()
	if dbViper.IsSet("blog.age") {
		age := dbViper.GetInt("blog.age")
		fmt.Println("age", age)
	} else {
		fmt.Println("blog.age不存在")
	}

	port := dbViper.GetInt("blog.port")
	fmt.Println("port", port)

	logViper := io.InitViper("../conf", "log", "yaml")
	type LogConfig struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	}
	var config LogConfig
	if err := logViper.Unmarshal(&config); err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		fmt.Println(config.Level)
		fmt.Println(config.File)
	}
}
