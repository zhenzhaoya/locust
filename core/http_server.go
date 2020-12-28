package core

import (
	"encoding/json"
	"fmt"
	"log"

	"io"
	"os"

	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zhenzhaoya/locust/utils"
)

var (
	_server  *HttpServer
	basePath string
)

type TaskData struct {
	Effect  bool
	Name    string
	Content string
}

type LoginData struct {
	Name     string
	Password string
}

func json2TaskData(b []byte) (*TaskData, error) {
	c := &TaskData{}
	err := json.Unmarshal(b, &c)
	return c, err
}

type UserContext struct {
	User    interface{}
	Code    int
	Message string
}

type LoginFunc func(string, string) (string, interface{})
type ContextFunc func(string, string, string) *UserContext
type GetFunc func() string
type SetFunc func([]byte) error
type SetFunc2 func(string) error
type HttpServer struct {
	IndexHtml  string
	Json       GetFunc
	Para       GetFunc
	Error      GetFunc
	Set        SetFunc
	HttpTask   SetFunc2
	login      LoginFunc
	user       ContextFunc
	httpFolder string
	port       int
}

type ResponseData struct {
	Code    int
	Message string
	Data    interface{}
}

func (res *ResponseData) ToJson() string {
	v, err := json.Marshal(res)
	if err == nil {
		return string(v)
	} else {
		return fmt.Sprintf(`{"Code":1000,"Message":"%v"}`, err.Error())
	}
}
func (res *ResponseData) Response(code int, message string) string {
	res.Code = code
	res.Message = message
	return res.ToJson()
}
func (res *ResponseData) Error(err error) string {
	if err == nil {
		return res.Success(nil)
	}
	res.Code = 1000
	res.Message = fmt.Sprintf(`%s`, err)
	return res.ToJson()
}
func (res *ResponseData) Success(data interface{}) string {
	res.Code = 0
	res.Message = "success"
	res.Data = data
	return res.ToJson()
}
func GetSuccessResponse(data interface{}) string {
	res := ResponseData{}
	return res.Success(data)
}
func GetResponse(code int, message string) string {
	res := ResponseData{}
	return res.Response(code, message)
}
func GetErrorResponse(err error) string {
	res := ResponseData{}
	return res.Error(err)
}

func GetHttpServer(indexHtml string, httpFolder string, get GetFunc, para GetFunc, err GetFunc,
	set SetFunc, httpTask SetFunc2, login LoginFunc, user ContextFunc) *HttpServer {
	if _server == nil {
		start := strings.LastIndex(indexHtml, "/")
		if start > -1 {
			basePath = indexHtml[0 : start+1]
		}
		_server = &HttpServer{IndexHtml: indexHtml, Json: get, Para: para,
			Error: err, Set: set, HttpTask: httpTask, httpFolder: httpFolder, login: login, user: user}
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
		w.Header().Set("Content-type", "text/plain")
	}
}
func (server *HttpServer) checkPermission(w http.ResponseWriter, r *http.Request) bool {
	user := server.getUser(r)
	if user != nil && user.Code == 0 {
		return false
	}
	if user != nil {
		io.WriteString(w, GetResponse(user.Code, user.Message))
	} else {
		io.WriteString(w, GetResponse(403, "No permission."))
	}
	return true
}
func (server *HttpServer) apiHandler(w http.ResponseWriter, r *http.Request) {
	if server.checkPermission(w, r) {
		return
	}
	if r.Method == "GET" {
		w.Header().Set("Content-type", "application/json")
		io.WriteString(w, server.Json())
	} else if r.Method == "POST" {
		// if server.checkPermission(w, r) {
		// 	return
		// }
		server.setConfig(w, r)
	}
}
func (server *HttpServer) setConfig(w http.ResponseWriter, r *http.Request) {
	if server.checkPermission(w, r) {
		return
	}
	w.Header().Set("Content-type", "application/json")
	s, err := ioutil.ReadAll(r.Body)
	if err == nil {
		err = server.Set(s)
		if err == nil {
			io.WriteString(w, GetSuccessResponse(nil))
			return
		}
	}
	io.WriteString(w, GetErrorResponse(err))
}
func (server *HttpServer) errHandler(w http.ResponseWriter, r *http.Request) {
	// if server.checkPermission(w, r) {
	// 	return
	// }
	w.Header().Set("Content-type", "application/json")
	io.WriteString(w, server.Error())
}
func (server *HttpServer) paraHandler(w http.ResponseWriter, r *http.Request) {
	if server.checkPermission(w, r) {
		return
	}
	w.Header().Set("Content-type", "application/json")
	io.WriteString(w, server.Para())
}
func (server *HttpServer) staticHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/users.json") {
		w.Header().Set("Content-type", "text/plain")
		io.WriteString(w, "404")
		return
	}
	if path == "/" {
		path = server.IndexHtml
	} else if strings.HasPrefix(path, "/http/") {
		path = server.httpFolder + path[6:]
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
	}
	if err != nil {
		w.Header().Set("Content-type", "text/plain")
		_indexHtml = fmt.Sprintf(`%s`, err)
	}
	io.WriteString(w, _indexHtml)
}
func (server *HttpServer) taskHandler(w http.ResponseWriter, r *http.Request) {
	if server.checkPermission(w, r) {
		return
	}
	w.Header().Set("Content-type", "application/json")
	res := ResponseData{}
	if r.Method == "GET" {
		files, err := utils.GetFiles(server.httpFolder, "*.http")
		if err != nil {
			io.WriteString(w, res.Error(err))
			return
		} else {
			l := len(server.httpFolder)
			for i := 0; i < len(files); i++ {
				files[i] = files[i][l:]
			}
			io.WriteString(w, res.Success(files))
			return
		}
	} else if r.Method == "POST" {
		s, err := ioutil.ReadAll(r.Body)
		if err != nil {
			io.WriteString(w, res.Error(err))
			return
		} else {
			task, err := json2TaskData(s)
			if err == nil {
				name := server.httpFolder + task.Name
				if task.Content == "" && task.Effect {
					err = server.HttpTask(name)
					io.WriteString(w, res.Error(err))
					return
				}
				err = utils.SaveFile(name, task.Content)
				if err == nil {
					if task.Effect {
						err = server.HttpTask(name)
						if err == nil {
							io.WriteString(w, res.Success(nil))
							return
						}
					} else {
						io.WriteString(w, res.Success(nil))
						return
					}
				}
			}
			io.WriteString(w, res.Error(err))
			return
		}
	} else {
		io.WriteString(w, res.Response(500, fmt.Sprintf("Request method '%s' not supported", r.Method)))
	}
}
func (server *HttpServer) getUser(r *http.Request) *UserContext {
	c, err := r.Cookie("__sid")
	if err != nil {
		return nil
	}
	return server.user(c.Value, r.Method, r.URL.Path)
}

func (server *HttpServer) userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	user := server.getUser(r)
	if user != nil {
		if user.Code == 0 {
			io.WriteString(w, GetSuccessResponse(user.User))
		} else {
			io.WriteString(w, GetResponse(user.Code, user.Message))
		}
	} else {
		io.WriteString(w, GetResponse(403, "No permission."))
	}
}
func (server *HttpServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	b, err := ioutil.ReadAll(r.Body)
	if err == nil {
		c := &LoginData{}
		err := json.Unmarshal(b, &c)
		if err == nil {
			token, user := server.login(c.Name, c.Password)
			if user == nil {
				io.WriteString(w, GetResponse(403, "Wrong user name or password"))
			} else {
				if token != "" {
					cookie := http.Cookie{Name: "__sid", Value: token, Path: "/", MaxAge: 60 * 60 * 24 * 30}
					http.SetCookie(w, &cookie)
				}
				io.WriteString(w, GetSuccessResponse(user))
			}
			return
		}
	}
	io.WriteString(w, GetErrorResponse(err))
}
func (server *HttpServer) StartServer(port int) {
	http.HandleFunc("/static/", server.staticHandler)
	http.HandleFunc("/api", server.apiHandler)
	http.HandleFunc("/err", server.errHandler)
	http.HandleFunc("/para", server.paraHandler)
	http.HandleFunc("/task", server.taskHandler)
	http.HandleFunc("/user", server.userHandler)
	http.HandleFunc("/login", server.loginHandler)
	http.HandleFunc("/", server.staticHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		panic(err)
	}
}
