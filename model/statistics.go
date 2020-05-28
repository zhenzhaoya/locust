package model

import (
	// "fmt"
	"encoding/json"
	"sync"
	"time"
)

var AverageTimeInterval float64 = 15

type Statistics struct {
	ReqCount        int // SuccessCount + FailCount
	DurationTime    float64
	DurationATime   float64
	RespACount      float64
	ReqACount       float64
	SuccessCount    int
	FailCount       int
	LastTime        time.Time
	Stoped          int
	mu              *sync.Mutex
	tmpReqCount     int
	tmpRespCount    int
	tmpDurationTime float64
}

func GetStatistics() *Statistics {
	var statics = new(Statistics)
	statics.LastTime = time.Now()
	statics.mu = new(sync.Mutex)
	return statics
}

func (s *Statistics) Reset() {
	var mu = s.mu
	mu.Lock()
	defer mu.Unlock()
	s.ReqCount = 0
	s.DurationTime = 0
	s.DurationATime = 0
	s.RespACount = 0
	s.ReqACount = 0
	s.SuccessCount = 0
	s.FailCount = 0
	s.LastTime = time.Now()
	s.tmpRespCount = 0
	s.tmpReqCount = 0
	s.tmpDurationTime = 0
}

func (u *Statistics) ToString() string {
	if b, err := json.Marshal(u); err == nil {
		return string(b)
	}
	return ""
}
func (s *Statistics) SetRequest(count int) {
	var mu = s.mu
	mu.Lock()
	defer mu.Unlock()
	s.ReqCount += count
	s.tmpReqCount += count
}
func (s *Statistics) SetResult(duration time.Duration, success bool) {
	mu := s.mu
	mu.Lock()
	defer mu.Unlock()
	if success {
		s.tmpRespCount += 1
		s.tmpDurationTime += duration.Seconds()
		s.SuccessCount += 1
		s.DurationTime += duration.Seconds()
	} else {
		s.FailCount += 1
	}
	seconds := time.Since(s.LastTime).Seconds()
	if seconds > AverageTimeInterval {
		s.LastTime = time.Now()
		if s.tmpRespCount > 0 {
			s.DurationATime = s.tmpDurationTime / (float64)(s.tmpRespCount)
		} else {
			s.DurationATime = 0
		}
		s.RespACount = (float64)(s.tmpRespCount) / seconds
		s.ReqACount = (float64)(s.tmpReqCount) / seconds
		s.tmpRespCount = 0
		s.tmpReqCount = 0
		s.tmpDurationTime = 0
	}
}
