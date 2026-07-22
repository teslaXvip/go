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

func TestWalkDir(t *testing.T) {
	io.WalkDir("../data")
}

func TestSplitFile(t *testing.T) {
	io.SplitFile("../img/懒大王.jpg", "../img/图像分割", 4)
}

func TestMergeFile(t *testing.T) {
	io.MergeFile("../img/图像分割", "../img/图像合并.jpg")
}

func TestCompress(t *testing.T) {
	io.Compress("../img/懒大王.jpg", "../img/懒大王2.jpg.zip")
}

func TestDecompress(t *testing.T) {
	io.Decompress("../img/懒大王2.jpg.zip", "../data/懒大王.jpg")
}
