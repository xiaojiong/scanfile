package scanfile

import (
	"encoding/json"
	"util"
)

/* 单个文件扫描后返回对象 */
type FileRes struct {
	File      string
	Strings   util.StringArray
	Num       int
	EndTime   int64
	StartTime int64
	RunTime   int64
}

func InitFileRes(file string) *FileRes {
	p := new(FileRes)
	p.File = file
	return p
}

func (res *FileRes) Add(str string) {
	res.Num++
	res.Strings.Set(str)
}

func (res *FileRes) End() {
	//res.runTime = res.endTime - res.startTime
}

/* ScanFile 扫描后返回对象 */
type ScanResult struct {
	ResList []*FileRes
	Counter *Counter
}

func (result *ScanResult) AddCounter(counter *Counter) {
	result.Counter = counter
}

func (result *ScanResult) AddFileRes(res *FileRes) {
	result.ResList = append(result.ResList, res)
}

func (result *ScanResult) ToJson() string {
	r, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}
	return string(r)
}
