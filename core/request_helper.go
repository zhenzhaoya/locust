package core

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"locust/model"
	"strings"
	"time"
)

type Options struct {
	Method  string
	Header  map[string]string
	Data    string
	Cookies []*http.Cookie
}

func NewDefaultOptions(cookies []*http.Cookie) Options {
	header := make(map[string]string, 0)
	header["Accept"] = "application/json, text/plain, */*"
	return Options{"GET", header, "", cookies}
}
func NewDefaultPostOptions(data string, cookies []*http.Cookie) Options {
	header := make(map[string]string, 0)
	header["Accept"] = "application/json, text/plain, */*"
	header["Content-Type"] = "application/json"
	return Options{"POST", header, data, cookies}
}
func Req(url string, opts Options) *model.RespStruct {
	client := &http.Client{}
	req, err := http.NewRequest(opts.Method, url, strings.NewReader(opts.Data))
	if err != nil {
		fmt.Println("Req.0", err)
	}
	if opts.Cookies != nil {
		for i := range opts.Cookies {
			req.AddCookie(opts.Cookies[i])
		}
	}
	if opts.Header != nil {
		for k, v := range opts.Header {
			req.Header.Set(k, v)
		}
	}
	var tb time.Time
	var el time.Duration
	tb = time.Now()
	resp, err := client.Do(req)
	el = time.Since(tb)

	var header http.Header
	var body []byte
	var statusCode int
	if err == nil {
		defer resp.Body.Close()
		statusCode = resp.StatusCode
		header = resp.Header
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Req.1", err)
		}
	}
	req.Close = true
	cookies := resp.Cookies()

	// fmt.Println(cookies)
	// fmt.Println(string(body)) //fmt.Sprintf("%s", err)
	return &model.RespStruct{StatusCode: statusCode, Header: header, Body: body, Err: err, Duration: el, Cookies: cookies}
}
