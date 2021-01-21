package fchan

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/yireyun/go-fwrite"
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
	return fmt.Sprintf(`FileName:"%s", LineNO:%v, Mark:"%v", `+
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

type FileConfig struct {
	fwrite.FileConfig
}

func (c *FileConfig) Config() *fwrite.FileConfig {
	return &c.FileConfig
}

func (w *FileConfig) GetFileRename(fileName string) (fileRename string, err error) {
	if fileName == "" {
		return "", fmt.Errorf("get file rename, fileName is null")
	}
	if w.RenameSuffix == "" {
		return "", fmt.Errorf("get file name, renameSuffix is null")
	}
	if w.WriteSuffix == "" {
		return "", fmt.Errorf("get file name, rriteSuffix is null")
	}

	if w.RotateRenameSuffix && w.RenameSuffix == w.WriteSuffix &&
		strings.HasSuffix(fileName, w.RenameSuffix) {
		return fileName, nil
	}

	if strings.HasSuffix(fileName, w.WriteSuffix) {
		fileName = fileName[:len(fileName)-len(w.WriteSuffix)]
	} else {
		return fileName, fmt.Errorf("get file rename, suffix is error")
	}

	fileRename = fileName + w.RenameSuffix
	if _, err = os.Lstat(fileRename); err == nil {
		for num := 1; err == nil && num <= math.MaxInt16; num++ { //出现重名时增加序号
			fileRename = fmt.Sprintf("%s.%03d%s", fileName, num, w.RenameSuffix)
			_, err = os.Lstat(fileRename)
		}
	}
	if err == nil {
		err = fmt.Errorf("Cannot find free file full number:%s", fileName)
	} else {
		err = nil
	}
	return
}

//获取文件行数
//fileName	是输入文件名
func (c *FileConfig) GetFileLines(fileName string) (int64, error) {
	return c.FileConfig.GetFileLines(fileName)
}

//获取重命名文件名
//fileName  	是输入文件名
//fileRename	是输出重命名文件名
//err       	是输出错误信息
func (w *FileConfig) GetNewFileName() (string, error) {
	if w.FilePrefix == "" {
		return "", fmt.Errorf("get file name, filePrefix is null")
	}
	if w.WriteSuffix == "" {
		return "", fmt.Errorf("get file name, writeSuffix is null")
	}
	if w.RenameSuffix == "" {
		return "", fmt.Errorf("get file name, renameSuffix is null")
	}
	now := time.Now()
	for num := 1; num <= math.MaxInt16; num++ {
		curDate := now.Format("2006-01-02")
		fileName := fmt.Sprintf("%s.%s.%03d%s", w.FilePrefix, curDate, num, w.WriteSuffix)
		fileRename := fmt.Sprintf("%s.%s.%03d%s", w.FilePrefix, curDate, num, w.RenameSuffix)
		fileClean := fmt.Sprintf("%s.%s.%03d%s", w.FilePrefix, curDate, num, w.CleanSuffix)
		_, fileNameErr := os.Lstat(fileName)
		_, fileRenameErr := os.Lstat(fileRename)
		_, fileCleanErr := os.Lstat(fileClean)
		if w.RotateRename && fileNameErr == nil && w.FileName != fileName { //文件重命名
			newName, err := w.GetFileRename(fileName)
			if err == nil {
				err = os.Rename(fileName, newName)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Printf("\t%s rename '%s' error:%v\n", w.Name, fileName, err)
			}
		}

		if fileNameErr != nil && fileRenameErr != nil && fileCleanErr != nil {
			return fileName, nil
		}
	}
	return "", fmt.Errorf("Cannot find free file name number:%s", w.FilePrefix)
}

func trimMark(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
