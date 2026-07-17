package io

import (
	"fmt"
	"os"
)

func CreateFile(fileName string) {
	os.Remove(fileName)
	if file, err := os.Create(fileName); err != nil {
		fmt.Println(err)
	} else {
		defer file.Close()
		file.Chmod(0o666)
		fmt.Printf("fd=%d\n", file.Fd())
		file.WriteString("哈哈哈哈\n")
		info, _ := file.Stat()
		fmt.Printf("is dir %t\n", info.IsDir())
		fmt.Printf("modify time %s\n", info.ModTime())
		fmt.Printf("mode %v\n", info.Mode())
		fmt.Printf("file name %s\n", info.Name())
		fmt.Printf("size %dB\n", info.Size())
	}
}
