package locust

import (
	"fmt"
	"sync"
	"time"

	"github.com/zhenzhaoya/locust/config"
	"github.com/zhenzhaoya/locust/core"
	"github.com/zhenzhaoya/locust/model"
	"github.com/zhenzhaoya/locust/utils"

	"encoding/json"
	"strconv"
	"strings"
)

type StartFunc func() *model.ContextModel
type HandlerFunc func(*model.ContextModel) *model.ContextModel
type TaskInfo struct {
	Handler HandlerFunc
	Rate    int
}
type StartInfo struct {
	Handler HandlerFunc
	Key     string
}
type Locust struct {
	UserConfig      *config.Config
	WaitingDuration func(*model.ContextModel) time.Duration
	UserLogout      func(*model.ContextModel)
	StartDuration   func() time.Duration
	SelfPara        func() map[string]interface{}
	SetData         func(*config.Config)
	urlMap          map[string]*model.Statistics
	errorMap        map[string]int
	mu              *sync.Mutex
	tasks           map[string]*TaskInfo
	startTasks      map[int]*StartInfo
	taskRate        [][2]int
	maxRate         int
	taskKeys        []string
	taskLen         int
	realCount       int
	selfData        float64
	startTask       StartFunc
	startKey        string
	endTask         HandlerFunc
	runFlag         bool
	subFlag         bool
	lastTime        time.Time
	httpServer      *core.HttpServer
	indexHtml       string
	httpFilePath    string
}

func APP() *Locust {
	locust := &Locust{}
	locust.indexHtml = "locust/index.html"
	return locust._init()
}
func GetAPP(indexHtmlPath string, httpFilePath string) *Locust {
	locust := &Locust{}
	locust.indexHtml = indexHtmlPath
	if httpFilePath != "" {
		httpFilePath = strings.TrimRight(httpFilePath, "/")
		httpFilePath = strings.TrimRight(httpFilePath, "\\")
		locust.httpFilePath = httpFilePath + "/"
	}
	return locust._init()
}

func SetConfig(app *Locust, configFile string) *config.Config {
	myConfig := app.UserConfig
	if configFile != "" {
		configFile = utils.GetPath(configFile)
		c := config.New(configFile)
		if c.UserCount > 0 {
			myConfig.UserCount = c.UserCount
		}
		if c.MinWait > 0 {
			myConfig.MinWait = c.MinWait
		}
		if c.MaxWait > 0 {
			myConfig.MaxWait = c.MaxWait
		}
		if c.StartMinWait > 0 {
			myConfig.StartMinWait = c.StartMinWait
		}
		if c.StartMaxWait > 0 {
			myConfig.StartMaxWait = c.StartMaxWait
		}
		if c.BaseUrl != "" {
			myConfig.BaseUrl = c.BaseUrl
		}
		if c.HttpFile != "" {
			myConfig.HttpFile = c.HttpFile
		}
		if c.Port > 0 {
			myConfig.Port = c.Port
		}

		myConfig.UserCount = c.UserCount
		myConfig.SelfDataName = c.SelfDataName
		dic := make(map[string]interface{}, 0)
		dic["BaseUrl"] = c.BaseUrl
		if c.HttpFile != "" {
			Init(utils.GetPath(c.HttpFile), app, dic)
		}

		return c
	}
	return nil
}

func (locust *Locust) getConfig() string {
	return locust.UserConfig.ToString()
}

func (locust *Locust) _init() *Locust {
	locust.UserConfig = config.GetDefault()
	locust.httpServer = core.GetHttpServer(locust.indexHtml, locust.httpFilePath, locust.getJson,
		locust.getConfig, locust.getError, locust.setData, locust.setHttpTask)
	locust.lastTime = time.Now()
	locust.runFlag = true
	locust.WaitingDuration = locust.waitingTime
	locust.StartDuration = locust.startDuration
	locust.errorMap = make(map[string]int)
	locust.urlMap = make(map[string]*model.Statistics)
	locust.tasks = make(map[string]*TaskInfo)
	locust.startTasks = make(map[int]*StartInfo)
	locust.mu = new(sync.Mutex)
	return locust
}
func (locust *Locust) startDuration() time.Duration {
	myConfig := locust.UserConfig
	return time.Duration(utils.GetRandNum(myConfig.StartMinWait, myConfig.StartMaxWait)) * time.Millisecond
}
func (locust *Locust) waitingTime(*model.ContextModel) time.Duration {
	myConfig := locust.UserConfig
	return time.Duration(utils.GetRandNum(myConfig.MinWait, myConfig.MaxWait)) * time.Millisecond
}
func (locust *Locust) setHttpTask(path string) error {
	defer func() {
		if error := recover(); error != nil {
			fmt.Println("setHttpTask: ", error)
		}
	}()
	runFlag := locust.runFlag
	if runFlag {
		locust.runFlag = false
	}
	locust.UserConfig.HttpFile = path
	locust.tasks = make(map[string]*TaskInfo)
	locust.startTasks = make(map[int]*StartInfo)

	dic := make(map[string]interface{}, 0)
	dic["BaseUrl"] = locust.UserConfig.BaseUrl
	Init(path, locust, dic)
	locust.restart()
	locust.runFlag = runFlag
	return nil
}
func (locust *Locust) setData(b []byte) error {
	defer func() {
		if error := recover(); error != nil {
			fmt.Println("setData: ", error)
		}
	}()
	// myConfig := locust.UserConfig
	myConfig, err := config.Json2Config(b)
	if err != nil {
		return err
	}
	locust.UserConfig = myConfig
	if myConfig.UserCount >= locust.realCount {
		go locust.userThread(myConfig.UserCount - locust.realCount)
	} else if locust.runFlag {
		locust.runFlag = false
		locust.subFlag = true
		go locust.removeUserThread()
	}

	if myConfig.Reset {
		go locust.reset()
	}
	if locust.SetData != nil {
		locust.SetData(myConfig)
	}
	return nil
}
func (locust *Locust) reset() {
	locust.mu.Lock()
	defer locust.mu.Unlock()
	locust.errorMap = make(map[string]int)
	for _, v := range locust.urlMap {
		v.Reset()
	}
	// locust.realCount = 0
	locust.selfData = 0
}
func (locust *Locust) getError() string {
	locust.mu.Lock()
	defer locust.mu.Unlock()
	var build strings.Builder = strings.Builder{}
	build.WriteString(`{`)
	var i = 0
	var n = len(locust.errorMap)
	for k, v := range locust.errorMap {
		i += 1
		if b, err := json.Marshal(v); err == nil {
			build.WriteString(`"`)
			build.WriteString(k)
			build.WriteString(`":`)
			build.WriteString(string(b))
			if i < n {
				build.WriteString(",")
			}
		}
	}
	build.WriteString("}")
	return build.String()
}
func (locust *Locust) getJson() string {
	locust.mu.Lock()
	defer locust.mu.Unlock()
	var build strings.Builder = strings.Builder{}
	build.WriteString(`{"UserCount":`)
	build.WriteString(strconv.Itoa(locust.realCount))
	build.WriteString(",")
	var t time.Time
	for k, v := range locust.urlMap {
		if time.Since(v.LastTime).Seconds() > model.AverageTimeInterval*6 {
			v.Stoped = 1
		}
		if b, err := json.Marshal(v); err == nil {
			build.WriteString(`"`)
			build.WriteString(k)
			build.WriteString(`":`)
			build.WriteString(string(b))
			if v.LastTime.After(t) {
				t = v.LastTime
			}
			build.WriteString(",")
		} else {
			fmt.Println(v.ToString())
			fmt.Println("json.Marshal", err)
		}
	}
	build.WriteString(fmt.Sprintf(`"SelfDataName":"%s",`, locust.UserConfig.SelfDataName))
	build.WriteString(fmt.Sprintf(`"SelfData":%f,`, locust.selfData))
	build.WriteString(fmt.Sprintf(`"RealCount":%d,`, locust.realCount))
	// if locust.WebSocketInfo != nil {
	// 	build.WriteString(fmt.Sprintf(`"WebSocket":%d,`, locust.WebSocketInfo()))
	// } else {
	// 	build.WriteString(fmt.Sprintf(`"WebSocket":%d,`, locust.realCount))
	// }
	build.WriteString(fmt.Sprintf(`"LastTime":"%s"`, t.Format("2006-01-02 15:04:05")))
	build.WriteString("}")
	return build.String()
}
func (locust *Locust) SetSelfData(d float64) {
	locust.selfData = d
}
func (locust *Locust) AddTask(method string, url string, handler HandlerFunc, rate float32) {
	locust.tasks[method+"_"+url] = &TaskInfo{handler, int(rate * 1000)}
}
func (locust *Locust) SetStartTask(method string, url string, handler StartFunc) {
	locust.startKey = method + "_" + url
	locust.startTask = handler
}
func (locust *Locust) AddStartTask(method string, url string, handler HandlerFunc) {
	index := len(locust.startTasks)
	locust.startTasks[index] = &StartInfo{handler, method + "_" + url}
}
func (locust *Locust) SetEndTask(handler HandlerFunc) {
	locust.endTask = handler
}
func (locust *Locust) Start(port int) {
	go locust.start()
	locust.httpServer.StartServer(port)
}
func (locust *Locust) restart() {
	locust.taskLen = len(locust.tasks)
	locust.taskKeys = make([]string, locust.taskLen)
	locust.taskRate = make([][2]int, locust.taskLen)
	if locust.taskLen == 0 {
		return
	}
	i := 0
	for k := range locust.tasks {
		locust.taskKeys[i] = k
		i += 1
	}

	num := 0
	for j := range locust.taskKeys {
		k := locust.taskKeys[j]
		t := locust.tasks[k]
		locust.taskRate[j][0] = num
		num += t.Rate
		locust.taskRate[j][1] = num
	}
	locust.maxRate = num
	fmt.Println("Total users:", locust.UserConfig.UserCount)
}
func (locust *Locust) start() {
	locust.restart()
	locust.userThread(locust.UserConfig.UserCount)
	for {
		if time.Since(locust.lastTime) > time.Minute {
			locust.lastTime = time.Now()
			locust.userThread(locust.UserConfig.UserCount - locust.realCount)
		}
		time.Sleep(time.Duration(10 * time.Second))
	}
}
func (locust *Locust) removeUserThread() {
	for {
		if locust.UserConfig.UserCount < locust.realCount {
			time.Sleep(time.Duration(10 * time.Millisecond))
		} else {
			break
		}
	}
	locust.subFlag = false
	locust.runFlag = true
	go locust.userThread(locust.UserConfig.UserCount - locust.realCount)
}
func (locust *Locust) userThread(count int) {
	if !locust.runFlag {
		return
	}
	for i := 0; i < count; i++ {
		if locust.UserConfig.UserCount <= locust.realCount {
			break
		}
		go locust.doTask()
		time.Sleep(locust.StartDuration())
	}
}
func (locust *Locust) setRealCount(count int) {
	locust.mu.Lock()
	locust.realCount += count
	locust.mu.Unlock()
}
func (locust *Locust) doTask() {
	locust.setRealCount(1)
	var urlKey string
	defer func() {
		if error := recover(); error != nil {
			fmt.Println("url: ", urlKey)
			fmt.Println("doTask: ", error)
		}
		locust.setRealCount(-1)
	}()
	var user *model.ContextModel
	if locust.startTask != nil {
		urlKey = locust.startKey
		var s *model.Statistics = locust.getStatistics(locust.startKey)
		s.SetRequest(1)
		user = locust.startTask()
		s.SetResult(user.Duration, user.Err == "")
		if user.Err != "" {
			locust.addError(urlKey, user.Err)
			return
		}
	}
	if user == nil {
		user = &model.ContextModel{}
		user.DefaultModel()
	}
	if len(locust.startTasks) > 0 {
		for i := 0; i < len(locust.startTasks); i++ {
			task := locust.startTasks[i]
			k := task.Key
			v := task.Handler
			var ss = locust.getStatistics(k)
			ss.SetRequest(1)
			tmp := v(user)
			if tmp == nil {
				ss.SetRequest(-1)
				continue
			}
			ss.SetResult(user.Duration, user.Err == "")
			if user.Err != "" {
				locust.addError(k, user.Err)
				return
			}
			time.Sleep(locust.WaitingDuration(user))
		}
	}

	i := 0
	for {
		if !locust.runFlag {
			if !(locust.subFlag && locust.UserConfig.UserCount >= locust.realCount) {
				if locust.UserLogout != nil {
					locust.UserLogout(user)
				}
				break
			}
		}
		if i >= locust.maxRate {
			i = 0
		}
		k, v := locust.getTask(i, user)
		if v != nil {
			i += 1
			urlKey = k
			var ss = locust.getStatistics(k)
			ss.SetRequest(1)
			tmp := v(user)
			if tmp == nil {
				ss.SetRequest(-1)
				continue
			}
			ss.SetResult(user.Duration, user.Err == "")
			if user.Err != "" {
				locust.addError(k, user.Err)
			} else {
				v, err := user.D["SelfData"]
				if err {
					var d float64
					dd := utils.ParseFloat(utils.GetStringValue(v), d)
					locust.SetSelfData(dd.(float64))
					delete(user.D, "SelfData")
				}
			}
		}
		time.Sleep(locust.WaitingDuration(user))
	}
	// locust.setRealCount(-1)
}
func (locust *Locust) AddStatistics(key string, duration time.Duration, message string, count int) {
	if message != "" {
		key = key + "_" + message
	}
	var ss = locust.getStatistics(key)
	ss.SetRequest(count)
	if count > 0 {
		ss.SetResult(duration, true)
	}
}
func (locust *Locust) AddError(key string, err string) {
	var ss = locust.getStatistics(key)
	ss.SetResult(0, false)
	locust.addError(key, err)
}
func (locust *Locust) addError(key string, err string) {
	err = strings.Replace(err, "\"", "\\\"", -1)
	errKey := fmt.Sprintf("%s<%s", key, err)
	locust.mu.Lock()
	num := locust.errorMap[errKey]
	locust.errorMap[errKey] = num + 1
	locust.mu.Unlock()
}
func (locust *Locust) getTask(n int, user *model.ContextModel) (string, HandlerFunc) {
	if locust.taskLen == 0 {
		return "", nil
	}
	if user != nil {
		next := user.Next
		if next == "" {
			v, err := user.D["nextTask"]
			if err {
				next = v.(string)
				delete(user.D, "nextTask")
			}
		} else {
			user.Next = ""
		}
		if next != "" {
			t := locust.tasks[next]
			if t != nil {
				return next, locust.tasks[next].Handler
			}
		}
	}

	if locust.UserConfig.NextRandom {
		n = utils.GetRandNum(0, locust.maxRate)
	}
	var i int
	for i = range locust.taskRate {
		if locust.taskRate[i][1] > n && locust.taskRate[i][0] <= n {
			break
		}
	}
	key := locust.taskKeys[i]
	var t = locust.tasks[key]
	return key, t.Handler
}
func (locust *Locust) getStatistics(key string) *model.Statistics {
	var s, ok = locust.urlMap[key]
	if !ok {
		locust.mu.Lock()
		defer locust.mu.Unlock()
		s, ok = locust.urlMap[key]
		if !ok {
			s = model.GetStatistics()
			locust.urlMap[key] = s
		}
	}
	return s
}
