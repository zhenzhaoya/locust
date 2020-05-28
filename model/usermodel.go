package model

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ContextModel struct {
	Data     map[string]string
	D        map[string]interface{}
	I        interface{}
	Next     string
	Duration time.Duration
	Err      string
	Cookies  []*http.Cookie
}

func (context *ContextModel) DefaultModel() *ContextModel {
	context.Data = make(map[string]string)
	context.D = make(map[string]interface{})
	return context
}
func (context *ContextModel) SetResp(resp *RespStruct, statusCode int) *ContextModel {
	context.Duration = resp.Duration
	if resp.StatusCode != statusCode {
		if resp.Err != nil {
			err := fmt.Sprintf("%s", resp.Err)
			i := strings.LastIndex(err, ":")
			if i > 0 {
				err = err[i+1:]
			}
			context.Err = fmt.Sprintf("StatusCode: %d, Error: %s", resp.StatusCode, err)
		} else {
			context.Err = fmt.Sprintf("StatusCode: %d, Body: %s", resp.StatusCode, resp.Body)
		}
	} else {
		context.Err = ""
	}
	return context
}
func (context *ContextModel) ValueCheck(resp *RespStruct, v []string, errorMessage string) *ContextModel {
	context.SetResp(resp, 200)
	if context.Err == "" {
		body := string(resp.Body)
		err := ""
		for i := range v {
			value := fmt.Sprintf(`"%s"`, v[i])
			if strings.Index(body, value) < 0 {
				if errorMessage == "" {
					err += fmt.Sprintf(`%s not exist. `, v[i])
				} else {
					err = errorMessage
					break
				}
			}
		}
		context.Err = err
	}
	return context
}
func (context *ContextModel) JsonCheck(resp *RespStruct, kvs map[string]interface{}, errorMessage string) *ContextModel {
	context.SetResp(resp, 200)
	if context.Err == "" {
		var build strings.Builder = strings.Builder{}
		body := string(resp.Body)
		for k, v := range kvs {
			i := strings.Index(body, fmt.Sprintf(`"%s"`, k))
			if i < 0 {
				if errorMessage != "" {
					break
				}
				build.WriteString(k)
				build.WriteString(" not exist. ")
				continue
			}
			var value string
			if fmt.Sprintf("%T", v) == "string" {
				value = fmt.Sprintf(`"%s"`, v)
			} else {
				value = fmt.Sprintf(`%d`, v)
			}
			j := strings.Index(body[i+len(k)+2:], value)
			if j < 0 {
				if errorMessage != "" {
					break
				}
				build.WriteString(k)
				build.WriteString("'s value not as expected. ")
				continue
			}
			mid := strings.Trim(body[i:j], "")
			if mid != ":" {
				if errorMessage != "" {
					break
				}
				build.WriteString(k)
				build.WriteString("'s value not as expected. ")
				continue
			}
		}
		if errorMessage == "" {
			context.Err = build.String()
		} else {
			context.Err = errorMessage
		}
	}
	return context
}
func (context *ContextModel) SetData(duration time.Duration, err string) *ContextModel {
	context.Duration = duration
	context.Err = err
	return context
}
func (context *ContextModel) SetError(duration time.Duration, statusCode int, err error) *ContextModel {
	context.Duration = duration
	r := fmt.Sprintf("%s", err)
	i := strings.LastIndex(r, ":")
	if i > 0 {
		r = r[i+1:]
	}
	context.Err = fmt.Sprintf("StatusCode: %d, Error: %s", statusCode, r)
	return context
}
