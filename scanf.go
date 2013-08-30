package scanfile

import (
	"io"
	"os"
	"sync"
)

var LineFeed = byte('\n') //文本换行符标识
var BufSize = 1024 * 1024 // buf大小

func Scan(files []string, searchStr *string) string {

	var result ScanResult
	//计数器
	counter := InitCounter(10)

	//扫描结果输出通道
	out := make(chan *FileRes, 10)

	fileCount := len(files)

	for i := 0; i < fileCount; i++ {
		go ScanFile(files[i], searchStr, counter, out)
	}

	for i := 0; i < fileCount; i++ {
		result.AddFileRes(<-out)
	}

	result.AddCounter(counter)
	return result.ToJson()
}

/* 扫描文件 */
func ScanFile(fileName string, searchStr *string, counter *Counter, out chan *FileRes) {
	//文件 IO
	fileContentChan := fileRead(fileName, counter)

	fileRes := InitFileRes(fileName)

	//使用多路复用 wg防止线程泄漏
	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			for {
				if text, ok := <-fileContentChan; ok {
					if counter.IsMax() {
						//清空未读取channel
						clearFileContentChan(fileContentChan)
						break
					} else {
						if counter.IsMax() {
							break
						}
						rs := strScan(text, searchStr, counter)
						for i := 0; i < len(rs); i++ {
							fileRes.Add(rs[i])
						}
					}
				} else {
					break
				}
			}
			wg.Done()
		}()
	}
	fileRes.End()
	wg.Wait()
	out <- fileRes
}

/* 消费闲置channel */
func clearFileContentChan(c chan *string) {
	for {
		if _, ok := <-c; ok == false {
			break
		}
	}
}

/* 文件IO操作，字符流放入channel */
func fileRead(fileName string, counter *Counter) chan *string {
	fileContentChan := make(chan *string, 5)
	go func() {
		fh, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}

		//异常处理
		defer fh.Close()

		buf := make([]byte, BufSize)

		var start int64
		fh.Seek(start, 0)
		for {
			//超过计数器最大返回值 跳出程序
			if counter.IsMax() {
				break
			}
			n, err := fh.Read(buf)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n == 0 {
				break
			}

			l := lastByteIndex(buf, LineFeed)
			content := string(buf[0 : l+1])
			start += int64(l + 1)
			fh.Seek(start, 0)
			fileContentChan <- &content
		}
		close(fileContentChan)
	}()
	return fileContentChan
}

/* last byte in slice byte */
func lastByteIndex(s []byte, sep byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == sep {
			return i
		}
	}
	return -1
}
