package io

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func SysCall() {
	cmd_path, err := exec.LookPath("go")
	if err != nil {
		fmt.Println("could not found command go")
	}
	fmt.Printf("command go in path %s\n", cmd_path)

	cmd := exec.Command("go", "version")
	if output, err := cmd.Output(); err != nil {
		fmt.Println("got output failed", err)
	} else {
		fmt.Println(string(output))
	}

	// 利用 runtime.Caller 拿到当前 .go 文件路径，推导出脚本所在目录。
	_, curFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(curFile)
	scriptPath := filepath.Join(dir, "hello.py")
	cmd = exec.Command("python", scriptPath)
	if output, err := cmd.Output(); err != nil {
		fmt.Println("python execute failed", err)
	} else {
		fmt.Println(string(output))
	}

	cmd = exec.Command("rm", "../data/biz.log")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	//如果不需要获得命令的输出，直接调用cmd.Run()即可
	err = cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	} else {
		fmt.Println(out.String())
	}
}
