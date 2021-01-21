package fchan

import (
	"math"
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

	fileEof []byte

	tailEof []byte

	cfg *FileConfig
}

//创建只写文件记录器
//name	是记录器名称
func NewTxtFileWrite(name string) *TxtFileWrite {
	w := new(TxtFileWrite)
	w.lineHead = TxtLineHead
	w.lineTail = TxtLineTail
	w.lineMark = LineMark
	w.fileEof, w.tailEof = GetTxtLineEof(w.lineHead, w.lineTail)
	w.cfg = NewFileConfig()
	w.cfg.InitAsDefault(name)
	w.cfg.SetFileEof(nil) //暂不追加尾行
	w.cfg.RotateRenameSuffix = true //初始为 true
	w.cfg.CleanRenameSuffix = true  //初始为 true
	w.cfg.FileLock = true           //初始为 true
	w.InitFileWriter(name, w.cfg)
	return w
}

//初始化
//fileSync  	是输入是否同步写文件
//filePrefix	是输入文件前缀
//writeSuffix   是输入正在写文件后缀
//renameSuffix  是输入重命名文件后缀
//cleanSuffix	是输入清理文件名后缀
//rotate    	是输入是否自动分割
//dayend     	是输入是否文件日终
//zeroSize  	是输入是否新文件零尺寸
//maxLines   	是输入最大行数,最小为1行
//maxSize   	是输入最大尺寸,最小为1M
//cleaning     	是输入是否清理历史
//maxDays		是输入最大天数,最小为3天
func (w *TxtFileWrite) Init(fileSync bool, filePrefix string,
	writeSuffix, renameSuffix, cleanSuffix string,
	rotate, dayend, zeroSize bool, maxLines, maxSize int64,
	cleaning bool, maxDays int, lastFiler LastFiler) (string, error) {

	w.cfg.lastFile = lastFiler
	return w.FileWrite.Init(fileSync, filePrefix,
		writeSuffix, renameSuffix, cleanSuffix,
		rotate, dayend, false, zeroSize, maxLines, maxSize,
		cleaning, maxDays)
}

//初始化行标志和标记
//lineHead	是输入行头标志，如:"---"，长度不能小于1
//lineTail	是输入行尾标志，如:"==="，长度不能小于1
//markSize	是输入行尾预留尺寸，不能小于32
func (w *TxtFileWrite) InitLineMark(lineHead, lineTail string, markSize int) error {
	lineHead = strings.TrimSpace(lineHead)
	lineTail = strings.TrimSpace(lineTail)

	if len(lineHead) < 0 {
		return ErrLineHeadNil
	}
	if len(lineTail) < 0 {
		return ErrLineTailNil
	}
	if markSize < 32 {
		return errorf("line mark len is not less than 32")
	}
	if markSize > 128 {
		return errorf("line mark len is not great than 128")
	}
	w.lineMark = strings.Repeat(" ", markSize)
	w.lineHead = lineHead
	w.lineTail = lineTail
	w.fileEof, w.tailEof = GetTxtLineEof(w.lineHead, w.lineTail)
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
	if len(line.Line.Bytes()) > math.MaxInt32 {
		return errorf("line bytes len is not great than MaxInt32")
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

//关闭文件
//err   	   	是输出错误信息
func (w *TxtFileWrite) Close() (err error) {
	return w.FileWrite.Close()
}
