package scanfile

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int IndexStr(char *haystack, char *needle,unsigned int begin) {
	char *p = strstr(haystack+begin, needle);
   	if (p)
      return p - haystack;
   	return -1;
}

int IndexChar(char *haystack, char c,unsigned int begin) {
    char *p = haystack = haystack + begin;

	while(*p != '\0') {
		if(*p == c) {
      		return p - haystack;
		}
		++p;
  	}
  	return -1;
}

int LastIndexChar(char *haystack, char c,unsigned int begin) {
	int len = strlen(haystack);
	if(begin > 0) {
		if (begin > len) {
			return -1;
		}
	} else {
		begin    = len - 1;

	}

	haystack +=begin;
	while(1) {
		if(*haystack == c) {
      		return begin;
		}
		if(begin == 0) {
			return -1;
		}
		--haystack;
		--begin;
	}
	return -1;
}
*/
import "C"
import "unsafe"

func strScan(str *string, key *string, counter *Counter) []string {
	begin := 0
	CStr := C.CString(*str)
	Ckey := C.CString(*key)

	defer func() {
		C.free(unsafe.Pointer(CStr))
		C.free(unsafe.Pointer(Ckey))
	}()

	var res []string

	for {
		var index int = 0
		if index = int(C.IndexStr(CStr, Ckey, C.uint(begin))); index == -1 {
			break
		}
		var startIndex int = 0
		if index > 0 {
			if pos := int(C.LastIndexChar(CStr, '\n', C.uint(index))); pos != -1 {
				startIndex = pos + 1
			}
		}
		var endIndex int = len(*str)
		if pos := int(C.IndexChar(CStr, '\n', C.uint(index))); pos != -1 {
			endIndex = pos + index
		}
		begin = endIndex

		if counter.IsMax() {
			break
		}
		res = append(res, (*str)[startIndex:endIndex])
		counter.Add()
		if begin == len(*str) {
			break
		}
	}
	return res
}
