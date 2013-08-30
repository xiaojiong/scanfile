package scanfile

import (
	"os"
	"path/filepath"
)

func PathFiles(path string) []string {
	fileNames := make([]string, 0)

	//遍历文件夹并把文件或文件夹名称加入相应的slice
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == false {

			fileNames = append(fileNames, path)
		}
		return err
	})
	if err != nil {
		panic(err)
	}
	return fileNames
}
