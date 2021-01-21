package fchan

import (
	"encoding/binary"
	"math"
	"strings"

	"github.com/yireyun/go-fwrite"
)

//
// |  4byes   |   byte   |   byte   |   byte   |   byte   |  4byes   |   4byes  |
// |  uint32  |   byte   |   byte   |   byte   |   byte   |  uint32  |   uint32 |
// | LineSize |  Version | HeadSize | MarkSize |  space   |  LineNo  | DateSize |
// |----------|----------|----------|----------|----------|----------|----------|
// | BigEndian|                                           | BigEndian| BigEndian|
//

const (
	HeadBytes        = 4 * 4    //行头16字节
	FixSize          = 4        //固定长度4字节
	HeadSize         = 4 + 4    //包头长度8字节
	FixByte1Ver      = 1        //格式版本
	FixByte2HeadSize = HeadSize //HeadSize
	FixByte3MarkSize = 0        //MarkSize

)

//只写文件记录器
type BinFileWrite struct {
	fwrite.FileWrite

	//行标记长度
	lineMark string

	fileEof []byte

	tailEof []byte

	cfg *FileConfig
}

//创建只写文件记录器
//name	是记录器名称
func NewBinFileWrite(name string) *BinFileWrite {
	w := new(BinFileWrite)

	w.lineMark = LineMark
	w.fileEof, w.tailEof = GetBinLineEof()
	w.cfg = NewFileConfig()
	w.cfg.InitAsDefault(name)
	w.cfg.SetFileEof(nil)
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
func (w *BinFileWrite) Init(fileSync bool, filePrefix string,
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
//markSize	是输入行尾预留尺寸，不能小于32
func (w *BinFileWrite) InitLineMark(markSize int) error {
	if markSize < 32 {
		return errorf("line mark len is not less than 32")
	}
	if markSize > 128 {
		return errorf("line mark len is not great than 128")
	}
	w.lineMark = strings.Repeat(" ", markSize)
	w.fileEof, w.tailEof = GetBinLineEof()
	return nil
}

//写入数据
//line    		是输入保存数据
//fileName  	是输出文件名
//lineNo    	是输出文件行号
//err   	   	是输出错误信息
func (w *BinFileWrite) Write(line *FileLine) (err error) {
	line.Mark = strings.TrimSpace(line.Mark)
	if len(line.Mark) > len(w.lineMark) {
		return errorf("line mark len more than %v", len(w.lineMark))
	}
	if len(line.Line.Bytes()) > math.MaxInt32 {
		return errorf("line bytes len is not great than MaxInt32")
	}
	line.buff.Reset()
	in := line.Line.Bytes()
	//进行文件旋转检查
	line.FileName, line.LineNO = w.cfg.RotateCheck(&w.FileWrite, len(in))
	//写入二进制数据头           1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16
	var lineHeadByte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var markSize byte = byte(len(w.lineMark) + 2)
	var lineSize uint32 = FixSize + HeadSize + uint32(len(in)) + uint32(markSize)
	binary.BigEndian.PutUint32(lineHeadByte[:4], lineSize)
	lineHeadByte[4] = FixByte1Ver //版本
	lineHeadByte[5] = HeadSize    //HAED数据区长度
	lineHeadByte[6] = markSize    //MARK标志区长度
	lineHeadByte[7] = 0           //空闲
	binary.BigEndian.PutUint32(lineHeadByte[8:12], uint32(line.LineNO))
	binary.BigEndian.PutUint32(lineHeadByte[12:16], uint32(len(in)))
	line.buff.Write(lineHeadByte)
	//写入行内容
	line.buff.Write(in)
	//写入行尾
	mark := []byte(" " + w.lineMark + "\n")
	line.use = len(line.Mark)
	copy(mark[1:], line.Mark)
	line.buff.Write(mark)
	//最后写入数据
	_, err = w.cfg.MutexWriter(&w.FileWrite, line.buff.Bytes())
	return
}

//关闭文件
//err   	   	是输出错误信息
func (w *BinFileWrite) Close() (err error) {
	return w.FileWrite.Close()
}
