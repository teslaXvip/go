package main

import (
	"bufio"
	"fmt"
	myHttp "go_basic/http"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func HttpOvservation() {
	resp, err := http.Get("http://127.0.0.1:5678/obs?name=zx&age=19")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Printf("response proto: %s\n", resp.Proto)
	if major, minor, ok := http.ParseHTTPVersion(resp.Proto); ok {
		fmt.Printf("http major version %d,http minor version %d\n", major, minor)
	}
	fmt.Printf("response status: %s\n", resp.Status)
	fmt.Printf("response status code: %d\n", resp.StatusCode)
	for key, values := range resp.Header {
		fmt.Printf("%s: %v url\n", key, values)
		if key == "Date" {
			if t, err := http.ParseTime(values[0]); err == nil {
				fmt.Println("server time:", t.Format("2006-01-02 15:04:05"))
			}
		}
	}
	fmt.Println("response body:")
	io.Copy(os.Stdout, resp.Body)
	os.Stdout.WriteString("\n\n")
	resp.Body.Close()
}

func Get() {
	fmt.Println(strings.Repeat("*", 30) + "GET" + strings.Repeat("*", 30))
	resp, err := http.Get("http://127.0.0.1:5678/get?" + myHttp.EncodeUrlParams(map[string]string{
		"name": "zx vip", "age": "18",
	}))
	if err != nil {
		panic(err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("response status: %s\n", resp.Status)
		fmt.Println("response body:")
		if body, err := io.ReadAll(resp.Body); err == nil {
			fmt.Println(string(body))
		}
		os.Stdout.WriteString("\n\n")
	}
}

func HugeBody() {
	fmt.Println(strings.Repeat("*", 30) + "GET HUGE BODY" + strings.Repeat("*", 30))
	if resp, err := http.Get("http://127.0.0.1:5678/stream"); err != nil {
		panic(err)
	} else {
		headerKey := http.CanonicalHeaderKey("Content-Length")
		fmt.Println("headerKey", headerKey)
		if ls, exists := resp.Header[headerKey]; exists {
			if l, err := strconv.Atoi(ls[0]); err == nil {
				haveRead := 0
				reader := bufio.NewReader(resp.Body)
				for {
					if bs, err := reader.ReadBytes('\n'); err == nil {
						haveRead += len(bs)
						progress := float64(haveRead) / float64(l)
						fmt.Printf("进度 %.2f%% ,内容 %s", progress*100, string(bs))
					} else {
						if err == io.EOF {
							if len(bs) > 0 {
								fmt.Println(string(bs))
							}
							break
						} else {
							fmt.Printf("read response body error: %s\n", err)
						}
					}
				}
			}
		}
	}
}

func main() {
	// HttpOvservation()
	// Get()
	HugeBody()
}
