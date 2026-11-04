package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsf/termbox-go"
)

func InitLogger() {
	if err := os.MkdirAll("log", os.ModePerm); err != nil {
		log.Printf("创建日志目录失败: %s", err)
		return
	}
	logFile, err := os.OpenFile("log/im_client.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Printf("打开日志文件失败: %s", err)
		return
	}
	log.SetOutput(logFile)
}

func ListenTermSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	log.Println("receive term signal " + sig.String() + ", going to exit")

	ReleaseWebSocket()
	if mq != nil {
		mq.Release()
	}
	CloseShell()

	os.Exit(0)
}

func main() {
	InitLogger()
	go ListenTermSignal()
	InitGroupBelong()
	InitApp()
	app.Run()

	// 正常退出（如exit命令）时释放资源
	ReleaseWebSocket()
	if mq != nil {
		mq.Release()
	}
	termbox.Close()
}

// 如果是Windows系统，最好在wsl里运行
// go run ./im/im_base_rabbitmq/client
// dqq-im » login 1
// dqq-im » send -t g1 -m 我是u1我来了
// Ctrl+A
// dqq-im » exit
