package model

import (
	"fmt"
	"net/http"
	"time"
)

type RespStruct struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Err        error
	Duration   time.Duration
	Cookies    []*http.Cookie
}

func (resp RespStruct) getErr() string {
	return fmt.Sprintf("%s", resp.Err)
}
