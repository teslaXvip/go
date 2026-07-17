package io_test

import (
	"fmt"
	"go_basic/io"
	"testing"
	"time"
)

func TestWriteFile(t *testing.T) {
	io.WriteFile()
}

func TestWriteFileWithBuffer(t *testing.T) {
	io.WriteFileWithBuffer()
}

func TestReadFile(t *testing.T) {
	io.ReadFile()
}

func TestReadFileWithBuffer(t *testing.T) {
	io.ReadFileWithBuffer()
}

func TestBufferedFileWriter(t *testing.T) {
	t1 := time.Now()
	io.WriteDirect("../data/no_buffer.txt")
	t2 := time.Now()
	io.WriteWithBuffer("../data/with_buffer.txt")
	t3 := time.Now()
	fmt.Printf("不用缓冲区耗时%dms, 用缓冲区耗时%dms\n", t2.Sub(t1).Milliseconds(), t3.Sub(t2).Milliseconds())
}

func TestCreateFile(t *testing.T) {
	io.CreateFile("../data/poem.txt")
}
