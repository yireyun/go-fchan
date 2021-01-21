package fchan

import (
	"strings"

	"github.com/yireyun/go-fwrite"
)

//只写文件记录器
type TxtFileWrite struct {
	fwrite.FileWrite

	//行头标志
	lineHead string

	//行尾标志
	lineTail string

	//行标记长度
	lineMark string

	cfg *FileConfig
}

//创建只写文件记录器
//fileName	是出文件名
//err   	是输出错误信息
func NewTxtFileWrite(name string) *TxtFileWrite {
	w := new(TxtFileWrite)

	w.lineHead = TxtLineHead
	w.lineTail = TxtLineTail
	w.lineMark = LineMark

	w.cfg = new(FileConfig)
	w.cfg.InitAsDefault(name)
	w.cfg.RotateRenameSuffix = true
	w.cfg.CleanRenameSuffix = true
	w.cfg.FileLock = true
	w.cfg.FileEof = nil
	w.InitFileWriter(name, w.cfg)
	return w
}

//初始化行标志和标记
//lineHead	是输入行头标志，如:"---"，长度不能小于1
//lineTail	是输入行尾标志，如:"==="，长度不能小于1
//markSize	是输入行尾预留尺寸，不能小于32
func (w *TxtFileWrite) InitLineMark(lineHead, lineTail string, markSize int) error {
	lineHead = strings.TrimSpace(lineHead)
	lineTail = strings.TrimSpace(lineTail)

	if len(lineHead) < 0 {
		return errorf("line head is null")
	}
	if len(lineTail) < 0 {
		return errorf("line Tail is null")
	}
	if markSize < 32 {
		return errorf("line mark len not less than 32")
	}

	w.lineHead = lineHead
	w.lineTail = lineTail
	w.lineMark = strings.Repeat(" ", markSize)
	return nil
}

//写入数据
//line    		是输入保存数据
//fileName  	是输出文件名
//lineNo    	是输出文件行号
//err   	   	是输出错误信息
func (w *TxtFileWrite) Write(line *FileLine) (err error) {
	line.Mark = strings.TrimSpace(line.Mark)
	if len(line.Mark) > len(w.lineMark) {
		return errorf("line mark len more than %v", len(w.lineMark))
	}

	line.buff.Reset()
	in := line.Line.Bytes()
	//检查没有换行符增加换行符
	if in[len(in)-1] != '\n' {
		line.Line.WriteByte('\n')
		in = line.Line.Bytes()
	}
	//进行文件旋转检查
	line.FileName, line.LineNO = w.cfg.RotateCheck(&w.FileWrite, len(in))
	//写入行头
	line.buff.WriteString(sprintf("%s%d\n", w.lineHead, line.LineNO))
	//写入行内容
	line.buff.Write(in)
	//写入行尾
	line.buff.WriteString(sprintf("%s%d;", w.lineTail, line.LineNO))
	mark := []byte(w.lineMark + "\n")
	line.use = len(line.Mark)
	copy(mark, line.Mark)
	line.buff.Write(mark)
	//最后写入数据
	_, err = w.cfg.MutexWriter(&w.FileWrite, line.buff.Bytes())
	return
}
