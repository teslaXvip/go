package main

import (
	"fmt"
	"io"
	"net/http"
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

func main() {
	http.HandleFunc("/obs", HttpObservation)
	if err := http.ListenAndServe("127.0.0.1:5678", nil); err != nil {
		panic(err)
	}
}
