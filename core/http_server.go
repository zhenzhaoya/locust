package core

import (
	"fmt"
	"log"

	"io"
	"os"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	_server  *HttpServer
	basePath string
)

type GetFunc func() string
type SetFunc func(map[string]interface{})
type HttpServer struct {
	IndexHtml string
	Json      GetFunc
	Para      GetFunc
	Error     GetFunc
	Set       SetFunc
	port      int
}

// func (server *HttpServer) init()  {
// 	server.indexHtml = "index.html"
// }
func GetHttpServer(indexHtml string, get GetFunc, para GetFunc, err GetFunc, set SetFunc) *HttpServer {
	if _server == nil {
		start := strings.LastIndex(indexHtml, "/")
		if start > -1 {
			basePath = indexHtml[0 : start+1]
		}
		_server = &HttpServer{IndexHtml: indexHtml, Json: get, Para: para, Error: err, Set: set}
	}
	return _server
}
func setContentType(w http.ResponseWriter, url string) {
	if strings.HasSuffix(url, ".css") {
		w.Header().Set("Content-type", "text/css")
	} else if strings.HasSuffix(url, ".js") {
		w.Header().Set("Content-type", "application/x-javascript")
	} else if strings.HasSuffix(url, ".htm") || strings.HasSuffix(url, ".html") {
		w.Header().Set("Content-type", "text/html")
	} else if strings.HasSuffix(url, ".json") {
		w.Header().Set("Content-type", "application/json")
	} else {
		w.Header().Set("Content-type", "text/html")
	}
}
func (server *HttpServer) apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	if r.Method == "GET" {
		io.WriteString(w, server.Json())
	} else if r.Method == "POST" {
		s, _ := ioutil.ReadAll(r.Body)
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(s), &dat); err == nil {
			fmt.Println("apiHandler: ", dat)
			server.Set(dat)
		}
		io.WriteString(w, `{"message":"ok"}`)
	}
}
func (server *HttpServer) errHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	io.WriteString(w, server.Error())
}
func (server *HttpServer) paraHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	io.WriteString(w, server.Para())
}
func (server *HttpServer) staticHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = server.IndexHtml
	} else {
		path = basePath + path[1:]
	}

	if strings.LastIndex(path, ".") < 0 || strings.LastIndex(path, ".") < len(path)-6 {
		path += ".html"
	}
	setContentType(w, path)
	var _indexHtml = ""
	f, err := os.Open(path)
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err == nil {
			_indexHtml = string(b)
		}
	} else {
		w.Header().Set("Content-type", "text/plain")
		_indexHtml = fmt.Sprintf(`%s`, err)
	}
	io.WriteString(w, _indexHtml)
}
func (server *HttpServer) StartServer(port int) {
	http.HandleFunc("/static/", server.staticHandler)
	http.HandleFunc("/api", server.apiHandler)
	http.HandleFunc("/err", server.errHandler)
	http.HandleFunc("/para", server.paraHandler)
	http.HandleFunc("/", server.staticHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		panic(err)
	}
}
