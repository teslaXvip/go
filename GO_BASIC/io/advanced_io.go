package io

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// func main() {
// 	fmt.Println("===== LimitReader 测试 =====")
// 	limitReader()

// 	fmt.Println("\n===== MultiReader 测试 =====")
// 	multiReader()

// 	fmt.Println("\n===== MultiWriter 测试 =====")
// 	multiWriter()

// 	fmt.Println("\n===== TeeReader 测试 =====")
// 	teeReader()

// 	fmt.Println("\n===== PipeIO 测试 =====")
// 	pipeIO()
// }

// LimitReader：限制最多读取N字节数据
func limitReader() {
	reader := strings.NewReader("daqiaoqiao")
	limitReader := io.LimitReader(reader, 6) // 最多读取6个字节
	content := make([]byte, 100)
	if n, err := limitReader.Read(content); err == nil {
		fmt.Printf("read %s\n", string(content[:n])) // daqiao
	}
	if _, err := limitReader.Read(content); err == io.EOF {
		fmt.Println("no more data available")
	}
}

// MultiReader：按顺序拼接多个Reader，依次读取
func multiReader() {
	r1 := strings.NewReader("黄梅时节家家雨\n")
	r2 := strings.NewReader("青草池塘处处蛙\n")
	r3 := strings.NewReader("有约不来过夜半\n")
	r4 := strings.NewReader("闲敲棋子落灯花\n")
	r := io.MultiReader(r1, r2, r3, r4) // 有序拼接多个读取源
	io.Copy(os.Stdout, r)
	// 应用场景：合并多个文件读取
}

// MultiWriter：一次写入，同步写入多个Writer
func multiWriter() {
	var (
		writer1 bytes.Buffer
		writer2 bytes.Buffer
	)
	multiWriter := io.MultiWriter(&writer1, &writer2)
	multiWriter.Write([]byte("黄梅时节家家雨\n"))

	fmt.Print(writer1.String()) // 黄梅时节家家雨
	fmt.Print(writer2.String()) // 黄梅时节家家雨
	// 应用场景：日志同时输出到控制台+文件
}

// TeeReader：读取源Reader数据时，自动复制一份写入指定Writer
func teeReader() {
	var writer bytes.Buffer
	reader := strings.NewReader("黄梅时节家家雨\n")
	teeReader := io.TeeReader(reader, &writer)
	io.Copy(os.Stdout, teeReader) // 读取teeReader，同时拷贝数据到writer
	fmt.Print(writer.String())
}

// Pipe：管道，goroutine间同步读写，无缓冲区
func pipeIO() {
	reader, writer := io.Pipe() // writer写入的数据直接流向reader
	go func() {
		writer.Write([]byte("hello"))
		writer.Close() // 写完必须关闭，否则读端阻塞
	}()

	content := make([]byte, 100)
	if n, err := reader.Read(content); err == nil {
		fmt.Printf("read %s\n", string(content[:n])) // hello
	}
	reader.Close()
}

/*
1. **LimitReader**：截断读取长度，防止一次性读取超大数据，常用于限制 HTTP 请求 body 大小
2. **MultiReader**：串联多个读取源，按顺序读完一个再读下一个，适合合并文件读取
3. **MultiWriter**：复制写入内容到多个输出对象，日志打印常用（同时打印控制台 + 落盘文件）
4. **TeeReader**：一边读取数据，一边自动备份一份，适合读取时同时备份原始流
5. **Pipe**：内存管道，跨 goroutine 同步读写，无缓冲区，写操作会阻塞直到读端接收数据
*/
