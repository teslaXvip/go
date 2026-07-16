package io_test

import (
	"go_basic/io"
	"testing"
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
