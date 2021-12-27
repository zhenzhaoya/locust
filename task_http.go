package locust

import (
	"github.com/zhenzhaoya/locust/core"
	"github.com/zhenzhaoya/locust/model"
	"github.com/zhenzhaoya/locust/utils"
)

func InitLines(lines []string, app *Locust, dic map[string]interface{}) bool {
	if lines != nil {
		if dic != nil {
			core.CopyConst(dic, core.ConstData)
		}
		core.DealJsonData(lines)
		tasks := core.NewHttpTask(lines)
		haveStart := false
		for i := 0; i < len(tasks); i++ {
			task := tasks[i]
			if task.StartTask {
				haveStart = true
				app.AddStartTask(task.Method, task.GetKey(), func(user *model.ContextModel) *model.ContextModel {
					if !haveStart {
						haveStart = true
						core.CopyConst(core.ConstData, user.D)
					}
					return httpTask(task, user)
				})
			} else {
				if task.GetKey() == "" {
					continue
				}
				app.AddTask(task.Method, task.GetKey(), func(user *model.ContextModel) *model.ContextModel {
					if !haveStart {
						haveStart = true
						core.CopyConst(core.ConstData, user.D)
					}
					return httpTask(task, user)
				}, task.Rate)
			}
		}
	}
	return true
}

func Init(path string, app *Locust, dic map[string]interface{}) bool {
	b := utils.PathExists(path)
	if !b {
		return false
	}
	lines := utils.GetFileLines(path)
	if lines == nil {
		panic("File not found: " + path)
	}
	return InitLines(lines, app, dic)
}

func httpTask(task *core.HttpTask, user *model.ContextModel) *model.ContextModel {
	d := task.GetKVPre(user.D)
	var url = task.GetUrl(d)
	var resp *model.RespStruct
	if task.Method == "GET" {
		opts := core.NewDefaultOptions(user.Cookies)
		for k, v := range task.GetHeader(d) {
			opts.Header[k] = v
		}
		resp = core.Req(url, opts)
	} else {
		var data = task.GetBody(d)
		// fmt.Println(data)
		opts := core.NewDefaultPostOptions(data, user.Cookies)
		for k, v := range task.GetHeader(d) {
			opts.Header[k] = v
		}
		opts.Method = task.Method
		resp = core.Req(url, opts)
	}
	if resp.Err != nil {
		return user.SetError(resp.Duration, resp.StatusCode, resp.Err)
	}
	err := task.CheckResult(resp.StatusCode, string(resp.Body), user.D)
	if resp.Cookies != nil && len(resp.Cookies) > 0 {
		user.Cookies = resp.Cookies
	}
	return user.SetData(resp.Duration, err)
}
