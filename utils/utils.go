package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetRandNum(min int, max int) int {
	if max <= min {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	var i int = rand.Intn(max-min) + min
	return i
}
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
func GetFileLines(path string) []string {
	f, err := ioutil.ReadFile(path)
	if err == nil {
		return strings.Split(string(f), "\n")
	}
	return nil
}
func ArrContains(arr []string, value string) bool {
	for k := range arr {
		if arr[k] == value {
			return true
		}
	}
	return false
}
func StrContains(value string, arr []string) bool {
	for k := range arr {
		if !strings.Contains(value, arr[k]) {
			return false
		}
	}
	return true
}
func StrContainsOr(value string, arr []string) bool {
	for k := range arr {
		if strings.Contains(value, arr[k]) {
			return true
		}
	}
	return false
}
func ArrContainsOr(arr []string, values []string) bool {
	for i := range arr {
		for j := range values {
			if arr[i] == values[j] {
				return true
			}
		}
	}
	return false
}
func Trim(str string) string {
	return strings.Trim(str, " ")
}

type JsonArrMap struct {
	Datas []map[string]interface{}
}
type JsonMap struct {
	Data map[string]interface{}
}

func Json2map(bs []byte) *JsonMap {
	j := &JsonMap{}
	var dat map[string]interface{}
	if err := json.Unmarshal(bs, &dat); err == nil {
		j.Data = dat
		return j
	}
	return nil
}
func (j *JsonMap) GetMap(key string) *JsonMap {
	v := j.Data[key]
	if v == nil {
		return nil
	}
	return &JsonMap{Data: v.(map[string]interface{})}
}
func (j *JsonMap) Get(key string) interface{} {
	return j.Data[key]
}
func (j *JsonMap) GetValue(key string) interface{} {
	ks := strings.Split(key, ".")
	l := len(ks)
	if l == 1 {
		return j.Data[ks[0]]
	} else {
		l = l - 1
		m := j.GetMap(ks[0])
		for i := 1; i < l; i++ {
			if m != nil {
				m = m.GetMap(ks[i])
			}
		}
		if m != nil {
			return m.Get(ks[l])
		} else {
			return nil
		}

	}
}
func (j *JsonMap) GetArrMap(key string) []map[string]interface{} {
	data := j.Data[key]
	if data == nil {
		return nil
	}
	dInterface := data.([]interface{})
	dMap := make([]map[string]interface{}, len(dInterface))
	for i := range dInterface {
		dMap[i] = dInterface[i].(map[string]interface{})
	}
	return dMap //data.([]map[string]interface{})
	// di := data.([]interface{})
	// dmap := make([]map[string]interface{},len(di))
	// for i := range di{
	// 	dmap[i]=di[i].(map[string]interface{})
	// }
	// return dmap
}
func (j *JsonMap) GetArr(key string) []interface{} {
	data := j.Data[key]
	if data == nil {
		return nil
	}
	return data.([]interface{})
}
func (j *JsonMap) GetStringArr(key string) []string {
	data := j.Data[key]
	if data == nil {
		return nil
	}
	return data.([]string)
}
func (j *JsonMap) GetString(key string) string {
	v := j.Data[key]
	if v == nil {
		return ""
	}
	return GetStringValue(v)
	// switch v.(type) {
	// case int:
	// case int64:
	// 	return fmt.Sprintf("%d", v)
	// case float64:
	// 	if v.(float64) == 0 {
	// 		return "0"
	// 	}
	// 	return fmt.Sprintf("%f", v)
	// }
	// return v.(string)
}
func GetValueType(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
func GetStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch v.(type) {
	case bool:
		return strconv.FormatBool(v.(bool))
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case time.Duration:
		return fmt.Sprintf("%d", v)
	case float64:
		if v.(float64) == 0 {
			return "0"
		}
		s := fmt.Sprintf("%f", v)
		if strings.Index(s, ".") > 0 {
			return strings.Trim(strings.Trim(s, "0"), ".")
		}
		return s
	}
	return v.(string)
}
func Atoi(s string, d int) int {
	i, err := strconv.Atoi(s)
	if err == nil {
		return i
	}
	return d
}
func ParseFloat(s string, d interface{}) interface{} {
	switch d.(type) {
	case float64:
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return float64(f)
		}
	case float32:
		f, err := strconv.ParseFloat(s, 32)
		if err == nil {
			return float32(f)
		}
	}
	return d
}
func GetLastSplit(s string, p string) string {
	i := strings.LastIndex(s, p)
	if i > 0 {
		s = s[i+len(p):]
	}
	return s
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
