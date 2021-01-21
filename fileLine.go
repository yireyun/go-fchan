package fchan

import (
	"bytes"
	"os"
)

//文件行
type FileLine struct {
	//文件名
	FileName string

	//文件已结束
	IsEof bool

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

	//读取文件
	readFD *os.File
}

func NewFileLine() *FileLine {
	l := new(FileLine)
	l.Line = bytes.NewBuffer(make([]byte, 0, LineSize))
	l.buff = bytes.NewBuffer(make([]byte, 0, LineSize+64+len(LineMark)))
	return l
}

func (l *FileLine) String() string {
	return sprintf(`{FileName:"%s", IsEof: %v, LineNO:%v, Mark:"%v", `+
		`off:%v, use:%v, free:%v, Line:"%s"}`,
		l.FileName, l.IsEof, l.LineNO, l.Mark, l.off, l.use, l.free,
		bytes.TrimSpace(l.Line.Bytes()))
}

func (l *FileLine) Clone(line *FileLine) {
	if l == nil || line == nil {
		return
	}
	l.Reset()
	l.FileName = line.FileName
	l.IsEof = line.IsEof
	l.Line.Write(line.Line.Bytes())
	l.LineNO = line.LineNO
	l.Mark = line.Mark
	l.buff.Write(line.buff.Bytes())
	l.off = line.off
	l.use = line.use
	l.free = line.free
	l.readFD = line.readFD
}

func (l *FileLine) Reset() {
	l.FileName = ""
	l.IsEof = false
	l.Line.Reset()
	l.LineNO = 0
	l.Mark = ""
	l.buff.Reset()
	l.off = 0
	l.use = 0
	l.free = 0
	l.readFD = nil
}
