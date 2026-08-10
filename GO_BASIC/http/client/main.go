package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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

func main() {
	HttpOvservation()
}
