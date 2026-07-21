package io

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func SplitFile(inFile string, outDir string, n int) error {
	if n <= 0 {
		return fmt.Errorf("n must be positive")
	}
	fin, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer fin.Close()

	stat, err := fin.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()
	chunk := fileSize / int64(n)
	if chunk <= 0 {
		return fmt.Errorf("file is too small or n is too large")
	}

	for i := range n {
		need := chunk
		if i == n-1 {
			need = fileSize - int64(n-1)*chunk // 全程 int64，避免溢出
		}
		fout, err := os.OpenFile(
			path.Join(outDir, strconv.Itoa(i)+"_"+path.Base(inFile)),
			os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		if _, err := io.CopyN(fout, fin, need); err != nil { // 保证读满
			fout.Close()
			return err
		}
		fout.Close()
	}
	return nil
}

func MergeFile(dir string, mergedFile string) error {
    fout, err := os.OpenFile(mergedFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer fout.Close()

    fileInfos, err := os.ReadDir(dir)
    if err != nil {
        return err
    }

    // TODO: 按 SplitFile 的命名规则做自然排序，避免 n>=10 时顺序错乱
    sort.Slice(fileInfos, func(i, j int) bool {
        return parseIdx(fileInfos[i].Name()) < parseIdx(fileInfos[j].Name())
    })

    for _, fi := range fileInfos {
        if !fi.Type().IsRegular() {
            continue
        }
        if err := AppendFile(fout, filepath.Join(dir, fi.Name())); err != nil {
            return err
        }
    }
    return nil
}

// parseIdx 解析文件名前缀中的序号，命名规则为 "{idx}_{baseName}"。
// 主要用于排序时实现自然排序：当文件数 ≥10 时，按字符串排序会出现
// "10_xxx" 排在 "2_xxx" 之前的问题，提取数值后比较可保证顺序正确。
// 若文件名不符合规则或解析失败，返回 0，使该文件排在排序结果的最前。
func parseIdx(name string) int {
	// 默认将整个 name 作为序号串，兼容没有 "_" 分隔符的情况
	idxStr := name
	// 找到首个 "_" 的位置，截取其前缀作为序号部分
	if i := strings.Index(name, "_"); i >= 0 {
		idxStr = name[:i]
	}
	// 将序号串转换为整数；转换失败时返回 0 以保证容错
	n, err := strconv.Atoi(idxStr)
	if err != nil {
		return 0
	}
	return n
}

func AppendFile(fout *os.File, infile string) error {
    fin, err := os.Open(infile)
    if err != nil {
        return err
    }
    defer fin.Close()

    _, err = io.Copy(fout, fin) // 自动处理 EOF、buffer、写错误
    return err
}

// func AppendFile(fout *os.File, infile string) {
// 	fin, err := os.Open(infile)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	defer fin.Close()

// 	buffer := make([]byte, 1024)
// 	for {
// 		n, err := fin.Read(buffer)
// 		if err != nil {
// 			if err == io.EOF {
// 				if n > 0 {
// 					fout.Write(buffer[:n])
// 				}
// 			} else {
// 				log.Println(err)
// 			}
// 			break
// 		} else {
// 			fout.Write(buffer[:n])
// 		}
// 	}
// }
