package io_test

import (
	"go_frame/io"
	"testing"

	"go.uber.org/zap"
)

func TestZap(t *testing.T) {
	logger := io.InitZap("../log/zap.log")
	defer logger.Sync()

	logger.Debug("hello")
	logger.Info("hello", zap.Int("age", 18))
	logger.Error("hello", zap.Namespace("china"), zap.Int("age", 18))

	sugar := logger.Sugar()
	sugar.Infof("pi is %f", 3.14)
}

func TestZap1(t *testing.T) {
	logger := io.InitZap1("../log/zap.log")
	defer logger.Sync()

	logger.Debug("hello")
	logger.Info("hello", zap.Int("age", 18))
	logger.Error("hello", zap.Namespace("china"), zap.Int("age", 18))

	sugar := logger.Sugar()
	sugar.Infof("pi is %f", 3.14)
}
