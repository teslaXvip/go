package io

import (
	"bufio"
	"fmt"
	"os"
)

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
