package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/zhenzhaoya/locust/utils"
)

var (
	mu        *sync.Mutex            = new(sync.Mutex)
	ConstData map[string]interface{} = make(map[string]interface{}, 0)
)

type HttpTask struct {
	// ConstD    map[string]interface{}
	CheckD    map[string][2]string
	Method    string
	url       string
	key       string
	Rate      float32
	header    map[string]string
	body      string
	StartTask bool
	isJson    bool
	mu        *sync.Mutex

	// Code map[string]func() string
	// data map[string]interface{} //临时变量
}

func saveConst(d map[string]interface{}, k string, v interface{}) {
	mu.Lock()
	defer mu.Unlock()
	d[k] = v
}
func getConst(d map[string]interface{}, k string) (interface{}, bool) {
	mu.Lock()
	defer mu.Unlock()
	v, ok := d[k]
	return v, ok
}
func CopyConst(d map[string]interface{}, aim map[string]interface{}) {
	for k, v := range d {
		aim[k] = v
	}
}

func (task *HttpTask) CheckResult(statusCode int, html string, d map[string]interface{}) string {
	// fmt.Println(html)
	s, ok := task.CheckD["StatusCode"]
	b := false
	if ok {
		var v float64
		v = utils.ParseFloat(s[1], v).(float64)
		b = contrastInt(s[0], v, float64(statusCode))
	} else {
		b = statusCode == 200
	}
	if !b {
		return fmt.Sprint("Data mismatch: StatusCode = ", statusCode)
	}
	var jMap *utils.JsonMap
	if task.isJson { // json数组先不处理
		jMap = utils.Json2map([]byte(html))
	}
	for k, v := range task.CheckD {
		if k == "StatusCode" {
			continue
		}
		switch k {
		case "ContainsValue":
			vs := strings.Split(v[1][1:len(v[1])-1], ",")
			for i := range vs {
				v1 := vs[i]
				if !strings.Contains(html, strings.Trim(v1, "\"")) {
					return "not contains " + v1
				}
			}
		case "ContainsReg":
			vs := strings.Split(v[1][1:len(v[1])-1], ",")
			for i := range vs {
				v1 := vs[i]
				reg, err := regexp.Compile(v1)
				if err == nil {
					if !reg.MatchString(html) {
						return "not contains " + v1
					}
				}
			}
		default:
			if v[0] == "=" {
				if !strings.Contains(v[1], "Data[\"") {
					continue
				}
				tmp := task.getDataValue(v[1], jMap, d)
				if utils.GetStringValue(tmp) == "" {
					return v[1] + " is null"
				} else {
					// task.D[k] = tmp
					d[k] = tmp
				}
			} else {
				var tk interface{}
				if strings.Index(k, "Data[\"") >= 0 {
					tk = task.getDataValue(k, jMap, d)
				} else {
					tk = task.getValue(k, d)
					f, err := strconv.ParseFloat(tk.(string), 64)
					if err == nil {
						tk = float64(f)
					}
				}
				var tv interface{}
				if strings.Index(v[1], "Data[\"") >= 0 {
					tv = task.getDataValue(v[1], jMap, d)
				} else {
					tv = task.getValue(v[1], d)
					f, err := strconv.ParseFloat(tv.(string), 64)
					if err == nil {
						tv = float64(f)
					}
				}
				if utils.GetValueType(tk) == "string" || utils.GetValueType(tv) == "string" {
					ttk := utils.GetStringValue(tk)
					ttv := utils.GetStringValue(tv)
					if !contrastStr(v[0], ttk, ttv) {
						return "Data mismatch: " + ttk + v[0] + ttv
					}
				} else {
					if !contrastInt(v[0], tk.(float64), tv.(float64)) {
						ttk := utils.GetStringValue(tk)
						ttv := utils.GetStringValue(tv)
						return "Data mismatch: " + ttk + v[0] + ttv
					}
				}

			}
		}
	}
	return ""
}
func (task *HttpTask) getDataValue(value string, jMap *utils.JsonMap, d map[string]interface{}) interface{} {
	for {
		index := strings.Index(value, "Data[\"")
		if index >= 0 {
			v1 := value[:index]
			tmp := value[index+6:]
			index = strings.Index(tmp, "\"]")
			key := tmp[:index]
			v2 := tmp[index+2:]
			v := jMap.GetValue(key)
			if v1 == "" && v2 == "" && v != nil && utils.GetValueType(v) != "string" {
				return v
				// value = v1 + utils.GetStringValue(v) + v2
			} else {
				value = v1 + utils.GetStringValue(v) + v2
			}
		} else {
			break
		}
	}
	value = task.dealString(getOperator(value), d)
	return value
}

func contrastStr(opt string, v string, v2 string) bool {
	switch opt {
	case "==":
		return v == v2
	case ">=":
		return v >= v2
	case ">":
		return v > v2
	case "<=":
		return v <= v2
	case "<":
		return v < v2
	case "!=":
		return v != v2
	}
	return true
}
func contrastInt(opt string, v float64, v2 float64) bool {
	// var v float64
	// v = utils.ParseFloat(v1, v).(float64)
	// v, _ := strconv.Atoi(v1)
	switch opt {
	case "==":
		return v == v2
	case ">=":
		return v >= v2
	case ">":
		return v > v2
	case "<=":
		return v <= v2
	case "<":
		return v < v2
	case "!=":
		return v != v2
	}
	return true
}
func newHttpTask() *HttpTask {
	task := &HttpTask{}
	// task.ConstD = d
	task.CheckD = make(map[string][2]string, 0)
	task.Rate = 1
	task.header = make(map[string]string, 0)
	// task.Code = make(map[string]func() string, 0)
	return task
}
func ClearArr(arr []string) {
	for i := 0; i < len(arr); i++ {
		arr[i] = ""
	}
}
func NewHttpTask(lines []string) []*HttpTask {
	tasks := make([]*HttpTask, 0)
	task := newHttpTask()
	startLine := -1
	codeIndex := 0
	startCode := false
	// startBody := -1
	dealCode := false
	var codeLines []string = make([]string, len(lines)/2, len(lines))
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		line = utils.Trim(line)
		if startCode {
			line = utils.Trim(line)
			if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "# ") {
				line = ""
			}
			if strings.HasSuffix(line, "%>") {
				startCode = false
				line = utils.Trim(strings.Trim(line, "%>"))
			}
			if line != "" {
				codeLines[codeIndex] = line
				codeIndex += 1
			}
			continue
		}
		if task.key != "" {
			if strings.HasPrefix(line, "###") || strings.HasPrefix(line, "POST ") || strings.HasPrefix(line, "GET ") || strings.HasPrefix(line, "PUT ") || strings.HasPrefix(line, "HEAD ") || strings.HasPrefix(line, "PATCH ") || strings.HasPrefix(line, "DELETE ") {
				task.dealSetCode(codeLines)
				tasks = append(tasks, task)
				task = newHttpTask()
				dealCode = true
				startCode = false
				ClearArr(codeLines)
				codeIndex = 0
				startLine = -1
				// break
			}
		}
		if strings.HasPrefix(line, "###") {
			if task.key == "" {
				startLine = i
			}
		}

		if strings.HasPrefix(line, "POST ") || strings.HasPrefix(line, "GET ") || strings.HasPrefix(line, "PUT ") || strings.HasPrefix(line, "HEAD ") || strings.HasPrefix(line, "PATCH ") || strings.HasPrefix(line, "DELETE ") {
			index := strings.Index(line, " ")
			// ps := strings.Split(line, " ")
			task.Method = line[0:index]
			task.url = line[index+1:]
			if startLine >= 0 {
				key := strings.Trim(lines[startLine], "###")
				key = utils.Trim(key)
				if strings.HasPrefix(key, "START") {
					task.StartTask = true
					task.key = key[5:]
				} else {
					task.key = key
				}
			}
			ps := strings.Split(task.url, ", ")
			if len(ps) > 1 {
				task.url = ps[0]
				var r float32 = 1
				task.Rate = utils.ParseFloat(ps[1], r).(float32)
			}
			if task.key == "" {
				task.key = task.url
			}
			dealCode = false
			startLine = i
			continue
		}

		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "# ") {
			startLine = i
			continue
		}
		if task.key == "" {
			continue
		}

		if i == startLine+1 {
			if line != "" && !strings.HasPrefix(line, "<%") {
				startLine = i
				index := strings.Index(line, ":")
				if index > 0 {
					task.header[utils.Trim(line[0:index])] = utils.Trim(line[index+1:])
				}
			}
			if strings.HasPrefix(line, "<%") {
				startCode = true
				if len(line) > 3 {
					codeLines[codeIndex] = utils.Trim(line[2:])
					codeIndex += 1
				}
			}
			continue
		}
		if strings.HasPrefix(line, "<%") {
			startCode = true
			line = utils.Trim(line[2:])
			if line != "" {
				codeLines[codeIndex] = line
				codeIndex += 1
			}
			continue
		}
		if line == "" {
			continue
		}
		task.body += line
	}
	if !dealCode {
		task.dealSetCode(codeLines)
		tasks = append(tasks, task)
	}
	return tasks
}
func getOperator(line string) []string {
	line = utils.Trim(line)
	opts := make([]string, 0)
	start := 0
	i := 0
	arr1 := []string{" == ", " != ", " >= ", " <= "}
	arr2 := []string{" > ", " < ", " ! ", " + ", " - ", " * ", " / "}
	arr3 := []string{"(", ")"}
	l := len(line)
	for ; i < l; i++ {
		if i < (l-4) && utils.ArrContains(arr1, line[i:i+4]) {
			opts = append(opts, line[start:i])
			opts = append(opts, line[i+1:i+3])
			i += 4
		} else if i < (l-3) && utils.ArrContains(arr2, line[i:i+3]) {
			opts = append(opts, line[start:i])
			opts = append(opts, line[i+1:i+2])
			i += 3
		} else if i < (l-1) && utils.ArrContains(arr3, line[i:i+1]) {
			opts = append(opts, line[start:i])
			opts = append(opts, line[i:i+1])
			i += 1
		} else {
			continue
		}
		start = i
	}
	if start < i {
		opts = append(opts, line[start:])
	}

	return opts
}
func (task *HttpTask) dealOperator(opts []string, d map[string]interface{}) string {
	midOpts := make([]string, 0)
	lastOpt := ""
	value := ""
	iValue := 0.0
	bValue := false
	sValue := ""
	iFlag := utils.ArrContainsOr(opts, []string{"-", "/", "*"})
	bFlag := utils.ArrContainsOr(opts, []string{"!", "==", ">=", "<=", "!=", ">", "<", "||", "&&"})
	sFlag := !iFlag && !bFlag
	if sFlag {
		return task.dealString(opts, d)
	}
	for i := 0; i < len(opts); i++ {
		v := opts[i]
		switch {
		case v == ">" || v == "<" || v == "==" || v == ">=" || v == "<=" || v == "!=":
			midOpts = append(midOpts, value)
			midOpts = append(midOpts, v)
		case v == "+" || v == "-" || v == "*" || v == "/" || v == "!":
			lastOpt = v
		// case v == "(":
		// 	startParentheses = true
		// case v == ")":
		// 	startParentheses = false
		default:
			if sFlag {
				realV := task.getValue(v, d)
				if lastOpt == "+" {
					sValue += realV
				} else {
					sValue += v
				}
			} else if bFlag {
				switch lastOpt {
				case "!":
					realV := task.getValueB(v, d)
					bValue = !realV
				case ">":
					realV := task.getValueI(v, d)
					bValue = iValue > realV
				case "<":
					realV := task.getValueI(v, d)
					bValue = iValue < realV
				default:
					iValue = task.getValueI(v, d)
				}
			} else if iFlag {
				realV := task.getValueI(v, d)
				switch lastOpt {
				case "+":
					iValue += realV
				case "-":
					iValue -= realV
				case "/":
					iValue /= realV
				case "*":
					iValue *= realV
				default:
					iValue = realV
				}
			}

		}
	}
	if iFlag {
		return strconv.FormatFloat(iValue, 'E', -1, 64)
	} else if bFlag {
		return strconv.FormatBool(bValue)
	} else {
		return value
	}
}
func (task *HttpTask) dealString(opts []string, d map[string]interface{}) string {
	// lastOpt := ""
	value := ""
	i := 0
	ignoreStr := []string{"(", ")"}
	// flag := false
	for ; i < len(opts); i++ {
		v := opts[i]
		if v == "+" || utils.ArrContains(ignoreStr, v) {
			continue
		} else { // if lastOpt == "+" {
			value += task.getValue(v, d)
		}
	}
	// if !flag {
	// 	value = task.getValue(opts[0])
	// }
	return value
}
func (task *HttpTask) getValueB(key string, d map[string]interface{}) bool {
	// data := task.D
	v, err := d[key]
	if err {
		return v.(bool)
	} else {
		v, err := getConst(ConstData, key)
		if err {
			return v.(bool)
		} else {
			b, _ := strconv.ParseBool(key)
			return b
		}
	}
}
func (task *HttpTask) getValueI(key string, d map[string]interface{}) float64 {
	// data := task.D
	v, err := d[key]
	if err {
		return v.(float64)
	} else {
		v, err := getConst(ConstData, key)
		if err {
			return v.(float64)
		} else {
			f, _ := strconv.ParseFloat(key, 64)
			return f
		}
	}
}
func (task *HttpTask) getValue(key string, d map[string]interface{}) string {
	if strings.HasPrefix(key, "\"") || strings.HasSuffix(key, "\"") {
		return strings.Trim(key, "\"")
	}
	// var data map[string]interface{}
	// if key[0] > 64 && key[0] < 97 {
	// 	data = task.D
	// } else {
	// 	data = task.data
	// }
	v, ok := getConst(d, key)
	if ok {
		return v.(string)
	} else {
		v, ok := getConst(ConstData, key)
		if ok {
			return v.(string)
		}
		return key
	}
}

// func (task *HttpTask) setKeyValue(key string, value string) {
// 	if key[0] > 64 && key[0] < 97 {
// 		task.D[key] = value
// 	} else {
// 		task.data[key] = value
// 	}
// }
// func (task *HttpTask) getKeyValue(key string) interface{} {
// 	if key[0] > 64 && key[0] < 97 {
// 		return task.D[key]
// 	} else {
// 		return task.data[key]
// 	}
// }

/**
*
 */
func (task *HttpTask) dealSetCode(lines []string) {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "# ") {
			continue
		}
		// [==,!=,>=,<=,>,<,!,+,-,*,/]
		if line != "" { // config language
			index := strings.Index(line, " ")
			startStr := line[0:index]
			line = line[index+1:]
			index = strings.Index(line, " ")
			opt := line[0:index]
			line = line[index+1:]
			// line = preDeal(line)
			// if opt == "=" && !utils.StrContains(line, []string{"Data[", "RandomNum(", "RandomString("}) {
			// 	value := task.dealString(getOperator(line))
			// 	task.setKeyValue(startStr, postDeal(value))
			// } else {
			if opt == "=" && strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
				// task.ConstD[startStr] = strings.Trim(line, "\"")
				if startStr[0] > 64 && startStr[0] < 97 {
					saveConst(ConstData, startStr, strings.Trim(line, "\""))
					continue
				}
			}
			if !task.isJson && (strings.HasPrefix(startStr, "Data[") || strings.HasPrefix(line, "Data[")) {
				task.isJson = true
			}

			task.CheckD[startStr] = [2]string{opt, line}
			// }
		}
	}
	// task.repleaceValue()
}

func (task *HttpTask) GetKVPre(d map[string]interface{}) map[string]interface{} {
	tmp := make(map[string]interface{}, 0)
	for k, v := range d {
		tmp[k] = v
	}
	// nextTask, ok := getConst(ConstData, "nextTask")
	// if ok {
	// 	d["nextTask"] = postDeal(task.dealString(getOperator(preDeal(nextTask.(string))), d))
	// }
	for k, v := range task.CheckD {
		opt := v[0]
		line := v[1]
		line = preDeal(line)
		if opt == "=" && !utils.StrContains(line, []string{"Data["}) {
			value := postDeal(task.dealString(getOperator(line), d))
			// task.setKeyValue(startStr, postDeal(value))
			if k[0] > 64 && k[0] < 97 || k == "nextTask" {
				d[k] = value
				// saveConst(ConstData, k, value)
			}
			tmp[k] = value
		}
	}
	return tmp
}

func (task *HttpTask) GetKey() string {
	v := task.key
	for {
		index := strings.Index(v, "<%=")
		if index >= 0 {
			index2 := strings.Index(v, "%>")
			line := v[index+3 : index2]
			value := task.dealString(getOperator(line), ConstData)
			v = v[0:index] + postDeal(value) + v[index2+2:]
		} else {
			v = utils.Trim(v)
			break
		}
	}
	return v
}
func (task *HttpTask) GetHeader(d map[string]interface{}) map[string]string {
	kv := make(map[string]string, 0)
	for k, v := range task.header {
		for {
			index := strings.Index(v, "<%=")
			if index >= 0 {
				index2 := strings.Index(v, "%>")
				line := v[index+3 : index2]
				value := task.dealString(getOperator(line), d)
				v = v[0:index] + postDeal(value) + v[index2+2:]
			} else {
				kv[k] = utils.Trim(v)
				break
			}
		}
	}
	return kv
}
func (task *HttpTask) GetUrl(d map[string]interface{}) string {
	v := task.url
	for {
		index := strings.Index(v, "<%=")
		if index >= 0 {
			index2 := strings.Index(v, "%>")
			line := v[index+3 : index2]
			value := task.dealString(getOperator(line), d)
			v = v[0:index] + postDeal(value) + v[index2+2:]
		} else {
			return utils.Trim(v)
		}
	}
}
func (task *HttpTask) GetBody(d map[string]interface{}) string {
	v := task.body
	for {
		index := strings.Index(v, "<%=")
		if index >= 0 {
			index2 := strings.Index(v, "%>")
			line := v[index+3 : index2]
			value := task.dealString(getOperator(line), d)
			v = v[0:index] + postDeal(value) + v[index2+2:]
		} else {
			return utils.Trim(v)
		}
	}
}
func preDeal(value string) string {
	key := "RandomNum("
	l := len(key)
	for {
		index := strings.Index(value, key)
		if index >= 0 {
			tmp1 := value[0:index]
			tmp2 := value[index+l:]
			index = strings.Index(tmp2, ")")
			nums := strings.Split(tmp2[0:index], ",")
			num1, _ := strconv.Atoi(utils.Trim(nums[0]))
			num2, _ := strconv.Atoi(utils.Trim(nums[1]))
			value = tmp1 + strconv.Itoa(utils.GetRandNum(num1, num2)) + tmp2[index+1:]
		} else {
			break
		}
	}
	key = "RandomString("
	l = len(key)
	for {
		index := strings.Index(value, key)
		if index >= 0 {
			tmp1 := value[0:index]
			tmp2 := value[index:]
			tmp2 = strings.Replace(tmp2, "(", "_", 1)
			tmp2 = strings.Replace(tmp2, ")", "_", 1)
			value = tmp1 + tmp2
		} else {
			break
		}
	}
	return value
}
func postDeal(value string) string {
	key := "RandomString_"
	l := len(key)
	for {
		index := strings.Index(value, key)
		if index >= 0 {
			tmp1 := value[0:index]
			tmp2 := value[index+l:]
			index = strings.Index(tmp2, "_")
			num, _ := strconv.Atoi(tmp2[0:index])
			value = tmp1 + utils.GetRandomString(num)
			if len(tmp2) > (index + 1) {
				value += tmp2[index+1 : len(tmp2)-1]
			}
		} else {
			break
		}
	}
	return value
}
