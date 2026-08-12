package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	myHttp "go_basic/http"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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

func Post() {
	bs, _ := json.Marshal(map[string]string{"name": "zx vv", "age": "18"})
	if resp, err := http.Post("http://127.0.0.1:5678/post", "application/json", bytes.NewReader(bs)); err != nil {
		panic(err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("response status: %s\n", resp.Status)
		fmt.Println("response body:")
		io.Copy(os.Stdout, resp.Body)
		os.Stdout.WriteString("\n\n")
	}

	if resp, err := http.PostForm("http://127.0.0.1:5678/post", url.Values{"name": []string{"zx vvip"}, "age": []string{"20"}}); err != nil {
		panic(err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("response status: %s\n", resp.Status)
		fmt.Println("response body:")
		io.Copy(os.Stdout, resp.Body)
		os.Stdout.WriteString("\n\n")
	}
}

func Cookie() {
	fmt.Println(strings.Repeat("*", 30) + "Cookie" + strings.Repeat("*", 30))
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:5678/cookie", nil)
	if err != nil {
		panic(err)
	}
	request.Header.Add("user-agent", "Mozilla/5.0(x64)")
	request.Header.Add("user-role", "vip")

	request.AddCookie(&http.Cookie{
		Name:   "auth",
		Value:  "pass",
		Domain: "localhost",
		Path:   "/",
	})

	request.AddCookie(&http.Cookie{
		Name:  "money",
		Value: "100",
	})

	request.AddCookie(&http.Cookie{
		Name:  "money",
		Value: "800",
	})

	client := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	if resp, err := client.Do(request); err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		fmt.Println("response header:")
		for k, v := range resp.Header {
			fmt.Println(k, v)
		}

		// 其实可以直接通过resp.Cookies()获得*http.Cookie，没必要自己解析
		if values, exists := resp.Header["Set-Cookie"]; exists {
			for _, value := range values { // 一个value对应一个response cookie
				cookie, _ := http.ParseSetCookie(value)
				fmt.Println("Name:", cookie.Name)
				fmt.Println("Value:", cookie.Value)
				fmt.Println("Domain:", cookie.Domain)
				fmt.Println("MaxAge:", cookie.MaxAge)
				fmt.Println(strings.Repeat("-", 50))
			}
		}

		os.Stdout.WriteString("\n\n")
	}
}

func main() {
	// HttpOvservation()
	// Get()
	// HugeBody()
	// Post()
	Cookie()
}
