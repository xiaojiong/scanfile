package main

import (
	"github.com/xiaojiong/memcachep"
	"github.com/xiaojiong/scanfile"

	"fmt"
	"log"
	"net"
)

func main() {
	port := 11345
	ls, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log.Fatalf("Got an error:  %s", e)
	}
	memcachep.Listen(ls)
}

//初始化绑定处理程序
func init() {
	memcachep.BindAction(memcachep.GET, GetAction)

}

func GetAction(req *memcachep.MCRequest, res *memcachep.MCResponse) {
	res.Fatal = false
	files := scanfile.PathFiles("C:\\test")

	key := string("290747680")

	content := scanfile.Scan(files, &key)

	log.Println(content)
	res.Value = []byte(string(content))
}
