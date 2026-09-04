package io_test

import (
	"go_frame/io"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLogrus(t *testing.T) {
	logger := io.InitLogrus("../log/logrus.log", "info") //相对于_test.go文件的路径
	logger.Debug("this is debug log")
	logEntry := logger.WithFields(logrus.Fields{"name": "dqq", "age": 18}) //日志中携带一些额外的key‑value
	logEntry.Info("this is info log")
	logEntry.Warnf("this is warn log, float=%.3f", 3.14)     //格式化输出
	logger.Error("this is error log1", "this is error log2") //多个string直接拼接在一起

	// logger.Fatal("this is fatal log")                     //写完日志之后会调os.Exit(1)

	defer func() {
		recover()
	}()
	logger.Panic("this is panic log") //写完日志之后会调panic
}

// go test -v ./io -run=^TestLogrus$ -count=1
