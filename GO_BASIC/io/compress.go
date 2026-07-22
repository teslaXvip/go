package io

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

// func Copy(inFile, outFile string) {
// 	fin, err := os.Open(inFile)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	fout, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	io.Copy(fout, fin)

// 	fout.Close()
// 	fin.Close()
// }

func Compress(inFile, outFile string) {
	fin, err := os.Open(inFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	stat, _ := fin.Stat()
	fmt.Printf("压缩前文件大小 %dB\n", stat.Size())

	fout, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	writer := gzip.NewWriter(fout)

	io.Copy(writer, fin)

	writer.Close()
	fout.Close()
	fin.Close()
}

func Decompress(inFile, outFile string) {
	fin, err := os.Open(inFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	stat, _ := fin.Stat()
	fmt.Printf("压缩后文件大小 %dB\n", stat.Size())

	reader, _ := gzip.NewReader(fin)

	fout, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	io.Copy(fout, reader)
	reader.Close()

	fout.Close()
	fin.Close()
}

const (
	_ = iota
	GZIP
	ZLIP
)

func Compress2(inFile, outFile string, compressAlgo int) {
	fin, err := os.Open(inFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fin.Close()

	stat, err := fin.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("压缩前文件大小 %dB\n", stat.Size())

	fout, err := os.OpenFile(outFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()

	var writer io.WriteCloser
	switch compressAlgo {
	case GZIP:
		writer = gzip.NewWriter(fout)
	case ZLIP:
		writer = zlib.NewWriter(fout)
	default:
		fmt.Printf("不支持的压缩算法: %d\n", compressAlgo)
		return
	}
	defer writer.Close()

	if _, err := io.Copy(writer, fin); err != nil {
		fmt.Println(err)
		return
	}
}
// defer 按相反顺序关闭（writer → fout → fin），符合资源生命周期