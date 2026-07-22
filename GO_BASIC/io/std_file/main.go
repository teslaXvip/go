package main

import (
	"fmt"
	"os"
)

func main() {
	// fmt.Scanf() 在单测里不生效
	Scan()
}

func Scan() {
	fmt.Println("please input a line")
	content := make([]byte, 100)
	n, err := os.Stdin.Read(content)
	if err == nil {
		fmt.Println("you input:", string(content[:n]))
	}
}
