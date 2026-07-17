package io

import (
	"bufio"
	"fmt"
	"os"
)

const logText string = "黄梅时节家家雨，青草池塘处处蛙。有约不来过夜半，闲敲棋子落灯花。\n"

func WriteFile() {
	if fout, err := os.OpenFile("../data/test.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666); err != nil {
		fmt.Println("出错了", err)
	} else {
		defer fout.Close()
		fout.WriteString("哈哈哈	\n")
		fout.WriteString("不是哥们\n")
		fout.WriteString("\n")
	}
}

func WriteFileWithBuffer() {
	if fout, err := os.OpenFile("../data/test.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666); err != nil {
		fmt.Println("出错了", err)
	} else {
		defer fout.Close()
		writer := bufio.NewWriter(fout)

		writer.WriteString("666\n")
		writer.Write([]byte("无情\n"))
		writer.WriteString("无敌了\n")
		writer.Flush()
	}
}

// 直接写文件
func WriteDirect(outFile string) {
	fout, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	for i := 0; i < 100000; i++ {
		fout.WriteString(logText)
	}
}

// 带缓冲写文件
func WriteWithBuffer(outFile string) {
	fout, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	writer := NewBufferedFileWriter(fout, 4096)
	defer writer.Flush() // 最后，务必把缓冲区里残留的内容写入磁盘
	for i := 0; i < 100000; i++ {
		writer.WriteString(logText)
	}
}
