package io_test

import (
	"encoding/json"
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

func TestUserMarshal(t *testing.T) {
	birthTime, _ := time.ParseInLocation(io.MyDateFormat, "2000-01-01", time.Local)
	u := io.User{
		Name:  "张三",
		Birth: io.MyDate(birthTime),
	}

	data, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}
	fmt.Println("序列化结果：")
	fmt.Println(string(data))

	want := `{
  "name": "张三",
  "birth": "2000-01-01"
}`
	if string(data) != want {
		t.Errorf("MarshalIndent got %s, want %s", string(data), want)
	}
}

func TestUserUnmarshal(t *testing.T) {
	data := []byte(`{
  "name": "张三",
  "birth": "2000-01-01"
}`)

	var u io.User
	err := json.Unmarshal(data, &u)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	got := time.Time(u.Birth).Format(io.MyDateFormat)
	fmt.Println("反序列化后birth：", got)

	want := "2000-01-01"
	if got != want {
		t.Errorf("Unmarshal birth got %s, want %s", got, want)
	}
}

func TestSysCall(t *testing.T) {
	io.SysCall()
}
