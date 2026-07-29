package concurrence

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	READ_FILE_ROUTINE_COUNT = 15 // 读取文件协程数量
	PROCESS_ROUTINE_COUNT   = 5  // 处理每行数据协程数量
)

var (
	sum        int64
	fileList   = make(chan string, 100)  // 存放待读取文件路径
	lineBuffer = make(chan string, 1000) // 存放文件每行内容
	walkWg     sync.WaitGroup
	readWg     sync.WaitGroup
	processWg  sync.WaitGroup
)

// walkDir 递归遍历目录，将普通文件路径写入fileList通道
func walkDir(dir string) {
	defer walkWg.Done()
	err := filepath.Walk(dir, func(subPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 判断是否为普通文件
		if info.Mode().IsRegular() {
			fileList <- subPath // 下游消费慢会阻塞
		}
		return nil
	})
	if err != nil {
		fmt.Printf("遍历目录异常: %v\n", err)
	}
}

// readFile 从fileList取文件，逐行读取内容推入lineBuffer
func readFile() {
	defer readWg.Done()
	for {
		infile, ok := <-fileList
		if !ok {
			break // channel关闭，退出循环
		}
		fin, err := os.Open(infile)
		if err != nil {
			fmt.Printf("打开文件%s失败: %v\n", infile, err)
			continue
		}

		// 延迟关闭文件句柄
		func() {
			defer fin.Close()
			reader := bufio.NewReader(fin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						// 文件末尾剩余不为空的一行也要处理
						if len(line) > 0 {
							lineBuffer <- strings.TrimSpace(line)
						}
						break
					} else {
						fmt.Printf("读取文件%s行失败: %v\n", infile, err)
						break
					}
				} else {
					// 去除首尾空格、换行符
					lineBuffer <- strings.TrimSpace(line)
				}
			}
		}()
	}
}

// processLine 从lineBuffer读取每行，转数字累加求和（原子操作保证并发安全）
func processLine() {
	defer processWg.Done()
	for {
		line, ok := <-lineBuffer
		if !ok {
			break
		}
		i, err := strconv.Atoi(line)
		if err != nil {
			fmt.Printf("%s not number\n", line)
		} else {
			atomic.AddInt64(&sum, int64(i))
		}
	}
}

// DealMassFile 入口函数：流水线并发处理海量文件
func DealMassFile(dir string) {
	// 监控通道堆积数量，每秒打印一次
	go func() {
		tk := time.NewTicker(time.Second)
		// tk: `*time.Ticker` 定时器结构体，用于定时触发事件
		// tk.C：Ticker 内部导出的只读通道（`<-chan time.Time`），定时器每到间隔时间，就会往这个 channel 塞一个当前时间值
		defer tk.Stop()
		for range tk.C {
			fmt.Printf("堆积了%d个文件未处理，堆积了%d行内容未处理\n",
				len(fileList), len(lineBuffer))
		}
	}()

	// WaitGroup计数初始化
	walkWg.Add(1)
	readWg.Add(READ_FILE_ROUTINE_COUNT)
	processWg.Add(PROCESS_ROUTINE_COUNT)

	// 1. 启动目录遍历协程
	go walkDir(dir)

	// 2. 启动N个文件读取协程
	for i := 0; i < READ_FILE_ROUTINE_COUNT; i++ {
		go readFile()
	}

	// 3. 启动M个行处理协程
	for i := 0; i < PROCESS_ROUTINE_COUNT; i++ {
		go processLine()
	}

	// 执行等待与关闭流水线通道
	walkWg.Wait()   // 等待目录全部遍历完成
	close(fileList) // 关闭文件通道，通知读文件协程没有新文件了

	readWg.Wait()     // 等待所有文件读取完毕
	close(lineBuffer) // 关闭行通道，通知处理协程没有新行数据

	processWg.Wait() // 等待所有行处理完成

	// 输出最终累加结果
	fmt.Printf("sum=%d\n", sum)
}
