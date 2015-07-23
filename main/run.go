package main

import (
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/xiaojiong/memcachep"
	"log"
	"net"
	"runtime"
	"scanfile"
)

var ConfigServerPath string
var ConfigServerPort int
var mf *scanfile.MemFiles

func init() {
	fmt.Println("server Init.")

	runtime.GOMAXPROCS(8)
	memcachep.BindAction(memcachep.GET, GetAction)

	/* 获取配置文件信息 */
	ini, err := goconfig.LoadConfigFile("./scanfile.conf")
	if err != nil {
		panic(err)
	}

	ConfigServerPath, err = ini.GetValue("server", "path")
	if err != nil {
		panic("config not found server.path")
	}

	ConfigServerPort, err = ini.Int("server", "port")
	if err != nil {
		panic("config not found server.port")
	}
}

func main() {
	files := scanfile.PathFiles(ConfigServerPath)
	mf := scanfile.InitMemFiles(files)

	ls, e := net.Listen("tcp", fmt.Sprintf(":%d", ConfigServerPort))
	if e != nil {
		log.Fatalf("Got an error:  %s", e)
	}

	fmt.Println("server running.")

	memcachep.Listen(ls)
}

func GetAction(req *memcachep.MCRequest, res *memcachep.MCResponse) {
	res.Fatal = false
	key := req.Key
	content := scanfile.MemScan(mf, &key)
	res.Value = []byte(string(content))
}
