package io

import (
	"fmt"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZap(logFile string) *zap.Logger {
	logger, _ := zap.NewProduction() //生产环境(设置日志级别为Info，error会打印调用堆栈，会上报位置）
	return logger
}

func InitZap1(logFile string) *zap.Logger {
	// 1. 编码器配置初始化
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"                                                    // 默认是ts
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder                           // 指定level的显示样式
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000") // 格式化时间
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// 2. 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("打开日志文件失败: %v", err))
	}

	// 3. 创建core
	core := zapcore.NewCore(
		// zapcore.NewJSONEncoder(encoderConfig), //日志为json格式
		zapcore.NewConsoleEncoder(encoderConfig), //日志为console格式 (Field还是json格式)
		zapcore.AddSync(file),                    //指定输出到文件
		zapcore.InfoLevel,                        //设置最低级别
	)

	// 4. 构造logger对象
	logger := zap.New(
		core,
		zap.AddCaller(),                       //上报文件名和行号
		zap.AddStacktrace(zapcore.ErrorLevel), //error级别及其以上的日志打印调用堆栈
		zap.Hooks(func(e zapcore.Entry) error { //可以添加多个钩子
			if e.Level >= zapcore.ErrorLevel {
				fmt.Println(e.Message)
			}
			return nil
		}),
	)

	logger = logger.With(
		zap.Namespace("dqq"), //后续的Field都记录在此命名空间中
		//通过zap.String、zap.Int等显式指定类型; fmt.Printf之类的方法大量使用interface{}和反射,
		//性能损失不少
		zap.String("biz", "search"), //公共的Field
	)

	return logger
}

func InitZap2(logFile string) *zap.Logger {
	// 1. 编码器配置初始化
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"                                                    // 默认是ts
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder                           // 指定level的显示样式
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000") // 格式化时间
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,    //单位为M, 文件大小超过这么多就会切分
		MaxBackups: 5,     //保留旧文件的最大个数
		MaxAge:     30,    //保留旧文件的最大天数
		Compress:   false, //是否压缩/归档旧文件
	}

	// 3. 创建core
	core := zapcore.NewCore(
		// zapcore.NewJSONEncoder(encoderConfig), //日志为json格式
		zapcore.NewConsoleEncoder(encoderConfig), //日志为console格式 (Field还是json格式)
		zapcore.AddSync(lumberJackLogger),        //指定输出到文件
		zapcore.InfoLevel,                        //设置最低级别
	)

	// 4. 构造logger对象
	logger := zap.New(
		core,
		zap.AddCaller(),                       //上报文件名和行号
		zap.AddStacktrace(zapcore.ErrorLevel), //error级别及其以上的日志打印调用堆栈
		zap.Hooks(func(e zapcore.Entry) error { //可以添加多个钩子
			if e.Level >= zapcore.ErrorLevel {
				fmt.Println(e.Message)
			}
			return nil
		}),
	)

	logger = logger.With(
		zap.Namespace("dqq"), //后续的Field都记录在此命名空间中
		//通过zap.String、zap.Int等显式指定类型; fmt.Printf之类的方法大量使用interface{}和反射,
		//性能损失不少
		zap.String("biz", "search"), //公共的Field
	)

	return logger
}

func InitZap3(logFile, level string) *zap.Logger {
	config := zap.Config{
		Encoding: "console", //日志默认是json格式
		// Encoding:         "json",    //日志默认是json格式
		OutputPaths:      []string{"stdout", logFile}, //输出到标准输出和指定的文件
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]any{"biz": "search"}, //公共的Field
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "msg",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			// EncodeLevel: zapcore.CapitalColorLevelEncoder,
			// EncodeLevel: zapcore.LowercaseColorLevelEncoder,
			// EncodeLevel: zapcore.LowercaseLevelEncoder,
		},
	}

	//动态解析日志级别
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		config.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	case "panic":
		config.Level = zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		panic(fmt.Errorf("invalid log level %s", level))
	}

	logger, _ := config.Build()
	logger = logger.With(
		zap.Namespace("dqq"), //后续的Field都记录在此命名空间中
		//通过zap.String、zap.Int等显式指定类型; fmt.Printf之类的方法大量使用interface{}和反射,
		//性能损失不少
		zap.String("group", "game"), //公共的Field
	)
	return logger
}
