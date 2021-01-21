package fchan

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yireyun/go-fwrite"
)

var (
	LineSize = 1024 //最大行尺寸
	MarkSize = 64
	LineMark = strings.Repeat(" ", MarkSize)
	LinePool = sync.Pool{New: func() interface{} { return NewFileLine() }}
)

var (
	ErrQueueEmpty    = fmt.Errorf("Queue Is Empty")
	ErrQueuePushNil  = fmt.Errorf("Push Bytes Is Nil")
	ErrQueueMarkNil  = fmt.Errorf("File Line Is Nil")
	ErrLineNotMatch  = fmt.Errorf("File Line Is Not Equi")
	ErrLineMarkEmpty = fmt.Errorf("Line Mark Is Empty")
	ErrQueueScaning  = fmt.Errorf("Queue Is Scaning File")
	ErrQueueWaitRead = fmt.Errorf("Queue Is Waiting Read")
	ErrQueueWaitMark = fmt.Errorf("Queue Is Waiting Mark")
	ErrQueueIsInit   = fmt.Errorf("Queue Is Already Init")
	ErrFileInvalid   = fmt.Errorf("FileName Is Invalid")
	ErrFileNotOpen   = fmt.Errorf("Queue File Not Open")
	ErrLineHeadNil   = fmt.Errorf("Line Head Is Null")
	ErrLineTailNil   = fmt.Errorf("Line Tail Is Null")
	ErrCheckFileEof  = fmt.Errorf("Check File Eof Fail")
)

type fileInfo struct {
	fileName  string
	fileDate  string    //文件日期
	fileOrder string    //文件序号
	fileTime  time.Time //文件时间
	fileIndex int       //文件序号
	isWrite   bool
}

type FileWriter interface {
	//写入数据
	//line    		是输入保存数据
	//err   	   	是输出错误信息
	Write(line *FileLine) (err error)
}

type FileReader interface {
	//写入数据
	//line    		是输入保存数据
	//err   	   	是输出错误信息
	Read(line *FileLine) (err error)
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

func checkWriteFileEof(fileName string, tailEof, lineEofTail []byte) error {
	flag := os.O_RDONLY | os.O_APPEND | os.O_SYNC
	fd, err := os.OpenFile(fileName, flag, 0660)
	if err != nil {
		return err
	}
	defer fd.Close()

	fs, err := fd.Stat()
	if err != nil {
		return err
	}

	if fs.Size() == 0 {
		_, err = fd.Write(tailEof)
		return nil
	}

	buf := make([]byte, len(tailEof))
	if fs.Size() >= int64(len(tailEof)) {
		sz, err := fd.ReadAt(buf, fs.Size()-int64(len(tailEof)))
		if err != nil {
			return err
		}
		if sz != len(tailEof) {
			return ErrCheckFileEof
		}

		bufTail := buf[len(buf)-len(lineEofTail):]
		if bytes.Compare(bufTail, lineEofTail) == 0 {
			return nil
		}
	}
	_, err = fd.Write(tailEof)
	return err
}

func SetOutput(out io.Writer) {
	fwrite.SetOutput(out)
}
