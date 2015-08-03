package scanfile

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var LineFeed = byte('\n') //文本换行符标识
var BufSize = 1024 * 1024 // buf大小
var MaxResult = 10

func MemScan(mf *MemFiles, searchStr *string) string {
	var result ScanResult
	//计数器
	counter := InitCounter(MaxResult)

	//扫描结果输出通道
	out := make(chan *FileRes, 10)

	go func() {
		for _, memFc := range mf.MFile {
			go MemScanFile(memFc, searchStr, counter, out)
		}
	}()

	for i := 0; i < len(mf.MFile); i++ {
		result.AddFileRes(<-out)
	}

	result.AddCounter(counter)
	return result.ToJson()
}

/* 扫描文件 */
func MemScanFile(fc *fileContent, searchStr *string, counter *Counter, out chan *FileRes) {

	//文件 IO
	fileContentChan := make(chan *string, 10)

	go func() {
		for fsi := 0; fsi < fc.Size; fsi++ {
			if counter.IsMax() {
				close(fileContentChan)
				break

			}
			fileContentChan <- fc.getSegment(fsi).Content
			if fsi+1 == fc.Size {
				close(fileContentChan)
			}
		}

	}()

	go func() {
		fileRes := InitFileRes(fc.FileName)

		//使用多路复用 wg防止线程泄漏
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
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
		wg.Wait()
		out <- fileRes
	}()
}

func Scan(files []string, searchStr *string) string {

	var result ScanResult
	//计数器
	counter := InitCounter(MaxResult)

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
	for i := 0; i < 10; i++ {
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

type fileSegment struct {
	Index   int
	Content *string
}

func (self *fileSegment) toString() string {
	return *self.Content
}

type fileContent struct {
	FileName string
	Size     int
	Segment  map[int]*fileSegment
}

func newFileContent(fileName string) *fileContent {
	fc := new(fileContent)
	fc.FileName = fileName
	fc.Segment = make(map[int]*fileSegment)
	return fc
}

func (self *fileContent) addSegment(fs *fileSegment, index int) {
	self.Size++
	self.Segment[index] = fs
}

func (self *fileContent) getSegment(index int) *fileSegment {
	seg, has := self.Segment[index]
	if has {
		return seg
	} else {
		return nil
	}
}

func InitMemFileContent(fileName string) *fileContent {
	fileContentCh := IoFileRead(fileName)
	fc := newFileContent(fileName)

	index := 0
	for {
		if content, ok := <-fileContentCh; ok {
			s := new(fileSegment)
			s.Index++
			s.Content = content

			fc.addSegment(s, index)

			index++
		} else {
			break
		}
	}
	return fc
}

type MemFiles struct {
	MFile map[string]*fileContent
}

func (self *MemFiles) addFile(filePath string) {
	fc := InitMemFileContent(filePath)
	self.MFile[filePath] = fc
}

func InitMemFiles(files []string) *MemFiles {
	mf := new(MemFiles)
	mf.MFile = make(map[string]*fileContent)

	for _, v := range files {
		mf.addFile(v)
		fmt.Println("Load file ", v)
	}

	return mf
}

/* 文件IO操作，字符流放入channel */
func IoFileRead(fileName string) chan *string {
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
