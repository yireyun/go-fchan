package fchan

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

var (
	LineSize = 1024 //最大行尺寸
	MarkSize = 64
	LineMark = strings.Repeat(" ", MarkSize)
	LinePool = sync.Pool{New: func() interface{} { return NewFileLine() }}
)

type FileWriter interface {
	//写入数据
	//line    		是输入保存数据
	//fileName  	是输出文件名
	//lineNo    	是输出文件行号
	//err   	   	是输出错误信息
	Write(line *FileLine) (err error)
}

func trimMark(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}