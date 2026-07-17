package io

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
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
            need = fileSize - int64(n-1)*chunk  // 全程 int64，避免溢出
        }
        fout, err := os.OpenFile(
            path.Join(outDir, strconv.Itoa(i)+"_"+path.Base(inFile)),
            os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
        if err != nil {
            return err
        }
        if _, err := io.CopyN(fout, fin, need); err != nil {  // 保证读满
            fout.Close()
            return err
        }
        fout.Close()
    }
    return nil
}