package fchan

import (
	"bytes"
	"strings"
)

//文件行
type FileLine struct {
	//文件名
	FileName string

	//文件行
	Line *bytes.Buffer

	//文件行号
	LineNO int64

	//行标记
	Mark string

	//缓存
	buff *bytes.Buffer

	//文件写偏移
	off int64

	//已用长度
	use int

	//可用长度
	free int
}

func NewFileLine() *FileLine {
	l := new(FileLine)
	l.Line = bytes.NewBuffer(make([]byte, 0, LineSize))
	l.buff = bytes.NewBuffer(make([]byte, 0, LineSize+64+len(LineMark)))
	return l
}

func (l *FileLine) String() string {
	return sprintf(`FileName:"%s", LineNO:%v, Mark:"%v", `+
		`off:%v, use:%v, free:%v, Line:"%v"`,
		l.FileName, l.LineNO, l.Mark, l.off, l.use, l.free,
		strings.TrimSpace(string(l.Line.Bytes())))
}

func (l *FileLine) Reset() {
	l.FileName = ""
	l.Line.Reset()
	l.LineNO = 0
	l.Mark = ""
	l.buff.Reset()
	l.off = 0
	l.use = 0
	l.free = 0
}
