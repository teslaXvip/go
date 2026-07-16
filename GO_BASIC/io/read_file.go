package io

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadFile() {
	if fin, err := os.Open("../data/test.txt"); err != nil {
		fmt.Println("xx", err)
	} else {
		defer fin.Close()
		bs := make([]byte, 100)
		fin.Read(bs)
		fmt.Println(string(bs))

		fin.Seek(0, 0)
		fin.Read(bs)
		fmt.Println(string(bs))

		fin.Seek(0, 0) // 由于前面已经把文件读完了，所以这次从头开始读
		const BATCH = 10
		for {
			buffer := make([]byte, BATCH)
			n, err := fin.Read(buffer)
			if n > 0 {
				fmt.Println(buffer[0:n])
			}
			if err == io.EOF { // End of File
				break
			}
			fin.Seek(0, 1)
		}
	}
	/*
		- Seek(offset, whence) 的第二个参数 whence ：
		- 0 = io.SeekStart （从文件开头算）
		- 1 = io.SeekCurrent （从当前位置算）
		- 2 = io.SeekEnd （从文件末尾算）
		- Seek(0, 1) = 从当前位置偏移 0 字节 = 位置不变，相当于空操作
	*/
}

func ReadFileWithBuffer() {
	if fin, err := os.Open("../data/test.txt"); err != nil {
		fmt.Println(err)
	} else {
		defer fin.Close()
		reader := bufio.NewReader(fin)
		for {
			str, err := reader.ReadString('\n')

			if len(str) > 0 {
				str = strings.TrimRight(str, "\n")
				fmt.Println(str)
			}

			if err == io.EOF {
				break
			}

		}
	}
}
