package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultPort = 8383
	HeaderRequestUrl = "Go-Proxy-Request-Url"
)

var url = ""
var headerExclude = []string{HeaderRequestUrl}

func main() {
	var pn int
	flag.IntVar(&pn, "p", DefaultPort, "port number (default: 8080)")

	var u string
	flag.StringVar(&u, "url", "", "remote url (default: '', use 'Go-Proxy-Request-Url' in header)")

	flag.Parse()

	port := fmt.Sprintf(":%d", pn)
	url = u

	start(port)
}

func start(port string) {
	fmt.Println("********************** Proxy Starting **********************")
	http.HandleFunc("/", request)
	http.ListenAndServe(port, nil)
	fmt.Println("********************** Proxy Shutdown **********************")
}

func request(w http.ResponseWriter, r *http.Request) {
	if url == "" {
		url = r.Header.Get(HeaderRequestUrl)
	}

	if url == "" {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("error: no remote url provided"))
		return
	}

	remote := fmt.Sprintf("%s%s", url, r.URL.Path)
	fmt.Println(remote)

	// TODO: Exclude some headers
	req, err := http.NewRequest(r.Method, remote, r.Body)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("error: " + err.Error()))
		return
	}

	for name, h := range r.Header {
		if !excludeHeader(name) {
			for _, h := range h {
				req.Header.Set(name, h)
			}
		}
	}

	timeout := time.Duration(30 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("error: " + err.Error()))
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("error: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	_, _ = w.Write(body)
}

func excludeHeader(header string) bool {
	for _, h := range headerExclude {
		if h == header {
			return true
		}
	}

	return false
}