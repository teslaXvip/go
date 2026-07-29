package concurrence

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// FileProcessor 封装海量文件处理的全部状态，避免全局变量耦合，支持多次调用
type FileProcessor struct {
	dir              string        // 待处理的根目录
	readRoutineCount int           // 读取文件协程数量
	processCount     int           // 处理每行数据协程数量
	maxOpenFiles     int           // 最大同时打开文件句柄数，防止 too many open files
	fileList         chan string   // 存放待读取文件路径
	lineBuffer       chan string   // 存放文件每行内容
	fileSem          chan struct{} // 文件句柄信号量，限制同时打开的文件数
	sum              int64         // 累加结果
	walkWg           sync.WaitGroup
	readWg           sync.WaitGroup
	processWg        sync.WaitGroup
}

// NewFileProcessor 创建 FileProcessor 实例
// readRoutineCount/processCount 为 0 时使用默认值
func NewFileProcessor(dir string, readRoutineCount, processCount, maxOpenFiles int) *FileProcessor {
	if readRoutineCount <= 0 {
		readRoutineCount = 15
	}
	if processCount <= 0 {
		processCount = 5
	}
	if maxOpenFiles <= 0 {
		maxOpenFiles = 100
	}
	return &FileProcessor{
		dir:              dir,
		readRoutineCount: readRoutineCount,
		processCount:     processCount,
		maxOpenFiles:     maxOpenFiles,
		fileList:         make(chan string, 100),
		lineBuffer:       make(chan string, 1000),
		fileSem:          make(chan struct{}, maxOpenFiles),
	}
}

// walkDir 递归遍历目录，将普通文件路径写入 fileList 通道
// 优化：使用 filepath.WalkDir 替代 filepath.Walk，减少 os.Lstat 系统调用开销（Go 1.16+）
func (p *FileProcessor) walkDir() {
	defer p.walkWg.Done()
	err := filepath.WalkDir(p.dir, func(subPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			p.fileList <- subPath // 下游消费慢会阻塞
		}
		return nil
	})
	if err != nil {
		fmt.Printf("遍历目录异常: %v\n", err)
	}
}

// readFile 从 fileList 取文件，逐行读取内容推入 lineBuffer
// 优化1：使用 fileSem 信号量限制同时打开的文件数，避免海量小文件触发 too many open files
// 优化2：使用 bufio.Scanner 替代 bufio.NewReader+ReadString，代码更简洁，EOF 处理更优雅
// 优化3：去掉原来多余的匿名函数包裹，defer fin.Close() 直接放在循环体内即可
func (p *FileProcessor) readFile() {
	defer p.readWg.Done()
	for infile := range p.fileList { // channel 关闭后自动退出
		// 获取信号量，限制同时打开的文件数
		p.fileSem <- struct{}{}
		fin, err := os.Open(infile)
		if err != nil {
			fmt.Printf("打开文件%s失败: %v\n", infile, err)
			<-p.fileSem // 打开失败也要释放信号量
			continue
		}

		func() {
			defer fin.Close()
			defer func() { <-p.fileSem }() // 文件关闭后释放信号量

			scanner := bufio.NewScanner(fin)
			// 优化：增大 Scanner 缓冲区至 256KB，避免单行过长时 Scanner 报错（默认仅 64KB）
			scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if len(line) > 0 {
					p.lineBuffer <- line
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("读取文件%s失败: %v\n", infile, err)
			}
		}()
	}
}

/*
fileSem不会有泄露风险。原因：

1. fileSem 不需要关闭 — 它是作为计数信号量使用的，只做 send （获取）和 receive （释放），没有任何 goroutine 用 range 遍历它，也没有 select 等待它关闭。
2. 不会阻塞 — 每次打开文件前 send ，文件关闭后 receive ，是严格配对的。当所有 readFile goroutine 执行完毕时，所有获取的信号量都已释放，channel 中不会有残留的 goroutine 阻塞在上面。
3. GC 自动回收 — 当 FileProcessor 不再被引用时， fileSem 作为其字段会随对象一起被垃圾回收。
对比需要关闭的场景 ：代码中 fileList 和 lineBuffer 必须关闭，因为消费端用了 for ... range ，需要通过关闭 channel 来通知消费者退出。而 fileSem 没有这种消费者模式，所以不需要关闭。
*/

// processLine 从 lineBuffer 读取每行，转数字累加求和（原子操作保证并发安全）
func (p *FileProcessor) processLine() {
	defer p.processWg.Done()
	for line := range p.lineBuffer { // channel 关闭后自动退出
		i, err := strconv.Atoi(line)
		if err != nil {
			fmt.Printf("%s not number\n", line)
		} else {
			atomic.AddInt64(&p.sum, int64(i))
		}
	}
}

// DealMassFile 入口函数：流水线并发处理海量文件
// 优化1：所有状态封装在 FileProcessor 结构体中，避免全局变量，支持多次调用
// 优化2：readRoutineCount/processCount/maxOpenFiles 作为参数传入，可动态调整
// 优化3：监控协程通过 done channel 退出，避免 goroutine 泄漏
func DealMassFile(dir string) {
	DealMassFileWithOptions(dir, 0, 0, 0)
}

// DealMassFileWithOptions 带参数的海量文件处理入口
// readRoutineCount: 读取文件协程数量（0 使用默认值 15）
// processCount: 处理行数据协程数量（0 使用默认值 5）
// maxOpenFiles: 最大同时打开文件数（0 使用默认值 100）
func DealMassFileWithOptions(dir string, readRoutineCount, processCount, maxOpenFiles int) {
	p := NewFileProcessor(dir, readRoutineCount, processCount, maxOpenFiles)

	// 优化：监控协程通过 done channel 控制退出，避免 goroutine 泄漏
	done := make(chan struct{})
	go func() {
		tk := time.NewTicker(time.Second)
		// tk: `*time.Ticker` 定时器结构体，用于定时触发事件
		// tk.C：Ticker 内部导出的只读通道（`<-chan time.Time`），定时器每到间隔时间，就会往这个 channel 塞一个当前时间值
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				fmt.Printf("堆积了%d个文件未处理，堆积了%d行内容未处理\n",
					len(p.fileList), len(p.lineBuffer))
			case <-done:
				return
			}
		}
	}()

	// WaitGroup 计数初始化
	p.walkWg.Add(1)
	p.readWg.Add(p.readRoutineCount)
	p.processWg.Add(p.processCount)

	// 1. 启动目录遍历协程
	go p.walkDir()

	// 2. 启动 N 个文件读取协程
	for i := 0; i < p.readRoutineCount; i++ {
		go p.readFile()
	}

	// 3. 启动 M 个行处理协程
	for i := 0; i < p.processCount; i++ {
		go p.processLine()
	}

	// 执行等待与关闭流水线通道
	p.walkWg.Wait()   // 等待目录全部遍历完成
	close(p.fileList) // 关闭文件通道，通知读文件协程没有新文件了

	p.readWg.Wait()     // 等待所有文件读取完毕
	close(p.lineBuffer) // 关闭行通道，通知处理协程没有新行数据

	p.processWg.Wait() // 等待所有行处理完成

	// 通知监控协程退出
	close(done)

	// 输出最终累加结果
	fmt.Printf("sum=%d\n", p.sum)
}
