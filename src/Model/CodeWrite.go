package Model

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
)

type CodeWrite struct {
	buf bytes.Buffer
}

func (_this *CodeWrite) Println(format string, args ...interface{}) {
	_, err := fmt.Fprintf(&_this.buf, format, args...)
	if err != nil {
		return
	}
}
func (_this *CodeWrite) Write(path string) {
	err := ioutil.WriteFile(path, _this.buf.Bytes(), 0666)
	if err != nil {
		log.Println("失败:", path, err)
	}
}
