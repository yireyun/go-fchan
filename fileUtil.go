package fchan

import (
	"strings"
	"sync"
	"unicode"
)

var (
	LineSize = 1024 //最大行尺寸
	LineHead = "---"
	LineTail = "==="
	LineMark = strings.Repeat(" ", 64)
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
