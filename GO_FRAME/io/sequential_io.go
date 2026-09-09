package io

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	ARRAY_SIZE = 1e7 // //1e8会超出goroutine栈大小的使用限制
)

// `[ARRAY_SIZE]int{}` 是**数组**，不是切片，直接定义全局变量放在堆上；如果定义在函数内，`1e7` 大小会直接栈溢出，所以放在全局 var。

var (
	arr   = make([]byte, ARRAY_SIZE)
	index = [ARRAY_SIZE]int{}
)

func InitArray() {
	for i := 0; i < len(index); i++ {
		index[i] = rand.Intn(len(arr))
	}
}

func WriteRamSequentially() {
	t0 := time.Now()
	defer func() {
		fmt.Printf("顺序写内存%dms\n", time.Since(t0).Milliseconds())
	}()

	for i := 0; i < len(arr); i++ {
		arr[i] = byte(i)
	}
}

func ReadRamSequentially() {
	t0 := time.Now()
	defer func() {
		fmt.Printf("顺序读内存%dms\n", time.Since(t0).Milliseconds())
	}()

	for i := 0; i < len(arr); i++ {
		_ = arr[i]
	}
}

func ReadRamRandomly() {
	t0 := time.Now()
	defer func() {
		fmt.Printf("随机读内存%dms\n", time.Since(t0).Milliseconds())
	}()

	for _, i := range index {
		_ = arr[i]
	}
}

func WriteRamRandomly() {
	t0 := time.Now()
	defer func() {
		fmt.Printf("随机写内存%dms\n", time.Since(t0).Milliseconds())
	}()

	for _, i := range index {
		arr[i] = byte(i)
	}
}

func WriteDiskSequentially() {
	t0 := time.Now()
	defer func() {
		fmt.Printf("顺序写磁盘%dms\n", time.Since(t0).Milliseconds())
	}()

	fout, err := os.OpenFile("./data/arr.bin", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	fout.Write(arr)
	fout.Sync() //把OS缓存的内容刷入磁盘
	fout.Close()
}

func ReadDiskSequentially() {
	t0 := time.Now()
	defer func() {
		fmt.Printf("顺序读磁盘%dms\n", time.Since(t0).Milliseconds())
	}()

	fin, err := os.Open("./data/arr.bin")
	if err != nil {
		panic(err)
	}
	defer fin.Close()
	fin.Read(arr)
}
