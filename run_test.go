package scanfile

import (
	"fmt"

	"testing"
)

func Test_T1(t *testing.T) {
	files := PathFiles("C:\\test")
	s := "290747680"
	r := Scan(files, &s)

	fmt.Println(r)
}
