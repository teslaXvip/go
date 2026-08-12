package main

import (
	"encoding/json"
	"fmt"
	myHttp "go_basic/http"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HttpObservation(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("request mehtod: %s\n", r.Method)
	fmt.Printf("request host: %s\n", r.Host)
	fmt.Printf("request url: %s\n", r.URL)
	fmt.Printf("request proto: %s\n", r.Proto)
	for key, values := range r.Header {
		fmt.Printf("%s: %v url\n", key, values)
	}

	fmt.Println()
	fmt.Printf("request body: ")
	// io.Copy(os.Stdout, r.Body)

	if body, err := io.ReadAll(r.Body); err == nil {
		fmt.Println(string(body))
	}
	fmt.Println()

	w.Header().Add("tRAce-id", "3144198515")
	w.Header().Add("tRAce-num", "3144198515")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello boy\n"))
	fmt.Fprintf(w, "hello girl\n")
}

func Get(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	params := myHttp.ParseUrlParams(r.URL.RawQuery)
	fmt.Fprintf(w, "your name is %s , age is %s \n", params["name"], params["age"])

}

func HugeBody(w http.ResponseWriter, r *http.Request) {
	line := []byte("Heavy is the head who wears the crown.\n")
	const R = 10
	totalSzie := R * len(line)
	w.Header().Add("content-length", strconv.Itoa(totalSzie))
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "不支持flusher", http.StatusInternalServerError)
		return
	}
	for i := range R {
		if _, err := w.Write(line); err != nil {
			fmt.Printf("%d send error :%s\n", i, err)
			break
		} else {
			flusher.Flush()
			time.Sleep(time.Second)
		}
	}
	fmt.Println(strings.Repeat("*", 60))
}

func Post(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if ct, exists := r.Header["Content-Type"]; exists {
		switch ct[0] {
		case "text/plain":
			io.Copy(w, r.Body) //直接把请求体作为响应体
		case "application/json":
			body, err := io.ReadAll(r.Body)
			if err == nil {
				params := make(map[string]string, 10)
				if err := json.Unmarshal(body, &params); err == nil {
					fmt.Fprintf(w, "your name is %s, age is %s\n", params["name"], params["age"])
				}
			} else {
				fmt.Println("read request body error", err)
			}
		case "application/x-www-form-urlencoded":
			body, err := io.ReadAll(r.Body)
			if err == nil {
				params := myHttp.ParseUrlParams(string(body))
				fmt.Fprintf(w, "your name is %s, age is %s\n", params["name"], params["age"])
			} else {
				fmt.Println("read request body error", err)
			}
		}
	}
}

func Cookie(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request header:")
	for key, value := range r.Header {
		fmt.Println(key, value)
	}

	// 其实可以直接通过r.Cookies()获得*http.Cookie，没必要自己解析
	if values, exists := r.Header["Cookie"]; exists {
		cookies, _ := http.ParseCookie(values[0]) // 多个request cookie全在values[0]里
		fmt.Println("request cookie:")
		for _, cookie := range cookies {
			fmt.Printf("%s: %s\n", cookie.Name, cookie.Value)
		}
		fmt.Println(strings.Repeat("*", 60))
	}

	// Set‑Cookie
	expiration := time.Now().Add(30 * 24 * time.Hour)
	cookie1 := http.Cookie{Name: "csrftoken", Value: "abcd", Expires: expiration, Domain: "localhost", Path: "/"}
	cookie2 := http.Cookie{Name: "jwt", Value: "1234", Expires: expiration, Domain: "localhost", Path: "/"}
	// 可以返回多个cookie
	http.SetCookie(w, &cookie1)
	http.SetCookie(w, &cookie2)
}

func main() {
	// http.HandleFunc("/obs", HttpObservation)
	// http.HandleFunc("/get", Get)
	// http.HandleFunc("/stream", HugeBody)
	// http.HandleFunc("/post", Post)
	// http.HandleFunc("/cookie", Cookie)

	// if err := http.ListenAndServe("127.0.0.1:5678", nil); err != nil {
	// 	panic(err)
	// }

	// go1.22以后标准库也支持灵活的路由设置了
	mux := http.NewServeMux()

	mux.HandleFunc("GET /obs", func(w http.ResponseWriter, r *http.Request) {
		HttpObservation(w, r)
	})

	mux.HandleFunc("GET /get", func(w http.ResponseWriter, r *http.Request) {
		Get(w, r)
	})

	mux.HandleFunc("POST /post", func(w http.ResponseWriter, r *http.Request) {
		Post(w, r)
	})

	mux.HandleFunc("GET /stream", func(w http.ResponseWriter, r *http.Request) {
		HugeBody(w, r)
	})

	mux.HandleFunc("GET /cookie", func(w http.ResponseWriter, r *http.Request) {
		Cookie(w, r)
	})

	// restful风格参数
	mux.HandleFunc("GET /get/{name}/{age}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "your name is %s, age is %s\n", r.PathValue("name"), r.PathValue("age"))
	})

	// mux.HandleFunc("GET /student", func(w http.ResponseWriter, r *http.Request) {
	// 	Student(w, r)
	// })

	if err := http.ListenAndServe("127.0.0.1:5678", mux); err != nil {
		panic(err)
	}
}
