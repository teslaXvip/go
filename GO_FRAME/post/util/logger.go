package util

import (
	"log/slog"
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
)

func InitSlog(logFile string) {
	// 创建日志切割writer
	fout, err := rotatelogs.New(
		logFile+".%Y%m%d%H",          // 日志文件名模板，会拼接时间
		rotatelogs.WithLinkName(logFile), // 创建软链接指向最新日志文件
		rotatelogs.WithRotationTime(1*time.Hour), // 每1小时切割一次
		rotatelogs.WithMaxAge(7*24*time.Hour),    // 保留7天日志，过期自动删除
	)
	if err != nil {
		panic(err)
	}

	handler := slog.NewTextHandler(
		fout, // 日志输出到切割文件
		&slog.HandlerOptions{
			AddSource: true,        // 打印日志的文件名+行号
			Level:     slog.LevelInfo, // 最低日志级别，Info及以上才输出
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// 修改日志time字段的时间输出格式
				if a.Key == slog.TimeKey {
					t := a.Value.Time()
					a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000"))
				}
				return a
			},
		},
	)

	logger := slog.New(handler)
	slog.SetDefault(logger) // 设置为全局默认logger
}
