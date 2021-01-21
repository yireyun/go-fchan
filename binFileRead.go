package fchan

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"strings"

	"github.com/yireyun/go-flock"
	"github.com/yireyun/go-fwrite"
)

//文件记录器
type BinFileReadWrite struct {
	Name     string
	fileName string
	fileCfg  *FileConfig
	flock    flock.Flocker
	fd       *os.File
	reader   *bufio.Reader
	readOff  int64
	lineHead []byte
	lineTail []byte
	lineMark string
	lineEof  *FileLine
}

//创建文件记录器
//fileName	是出文件名
//err   	是输出错误信息
func NewBinFileReadWrite(name string) *BinFileReadWrite {
	rw := new(BinFileReadWrite)
	rw.Name = name
	rw.fileCfg = new(FileConfig)
	rw.lineHead = []byte(TxtLineHead)
	rw.lineTail = []byte(TxtLineTail)
	rw.lineMark = LineMark
	rw.lineEof = NewFileLine()
	return rw
}

//初始化
//fileSync  	是输入是否同步写文件
//filePrefix	是输入文件前缀
//writeSuffix   是输入正在写文件后缀
//renameSuffix  是输入重命名文件后缀
//cleanSuffix	是输入清理文件名后缀
//cleaning     	是输入是否清理历史
//maxDays		是输入最大天数,最小为3天
func (rw *BinFileReadWrite) Init(fileSync bool,
	filePrefix, writeSuffix, renameSuffix, cleanSuffix string,
	cleaning bool, maxDays int) (fileName string, err error) {

	prefix := func(s string) string {
		s = strings.TrimSpace(s)
		if l := len(s); l > 0 && s[l-1] == '.' {
			return s[:l-1]
		} else {
			return s
		}
	}

	suffix := func(s string) string {
		s = strings.TrimSpace(s)
		if l := len(s); l > 0 && s[0] != '.' {
			return "." + s
		} else {
			return s
		}
	}
	filePrefix = prefix(filePrefix)
	if filePrefix == "" {
		return "", errorf("filePrefix is null")
	}
	writeSuffix = suffix(writeSuffix)
	if writeSuffix == "" {
		return "", errorf("writeSuffix is null")
	}
	renameSuffix = suffix(renameSuffix)
	if renameSuffix == "" {
		return "", errorf("renameSuffix is null")
	}
	cleanSuffix = suffix(cleanSuffix)
	if cleanSuffix == "" {
		return "", errorf("cleanSuffix is null")
	}

	if cleaning && maxDays < fwrite.MaxKeepDays { //最小为3天
		return "", errorf("maxDays not less than 3 day")
	}

	if rw.fileCfg.FilePrefix == filePrefix &&
		rw.fileCfg.WriteSuffix == writeSuffix &&
		rw.fileCfg.RenameSuffix == renameSuffix &&
		rw.fileCfg.CleanSuffix == cleanSuffix &&
		rw.fileCfg.Cleaning == cleaning &&
		rw.fileCfg.MaxDays == maxDays {
		return rw.fileCfg.FileName, nil
	}

	rw.fileCfg.FilePrefix = filePrefix
	rw.fileCfg.WriteSuffix = writeSuffix
	rw.fileCfg.RenameSuffix = renameSuffix
	rw.fileCfg.CleanSuffix = cleanSuffix
	if rw.fileCfg.RotateRenameSuffix {
		rw.fileCfg.RotateRename = writeSuffix != renameSuffix
	} else {
		rw.fileCfg.RotateRename = true
	}
	if rw.fileCfg.CleanRenameSuffix {
		rw.fileCfg.CleanRename = writeSuffix != renameSuffix
	} else {
		rw.fileCfg.CleanRename = false
	}
	rw.fileCfg.MaxDays = maxDays

	rw.fileCfg.FileName = rw.fileCfg.FilePrefix + rw.fileCfg.WriteSuffix
	return rw.fileCfg.FileName, nil
}

//初始化行标志和标记
//lineHead	是输入行头标志，如:"---"，长度不能小于1
//lineTail	是输入行尾标志，如:"==="，长度不能小于1
func (rw *BinFileReadWrite) InitLineMark(lineHead, lineTail string) error {
	lineHead = strings.TrimSpace(lineHead)
	lineTail = strings.TrimSpace(lineTail)

	if len(lineHead) < 0 {
		return ErrLineHeadNil
	}
	if len(lineTail) < 0 {
		return ErrLineTailNil
	}

	rw.lineHead = []byte(lineHead)
	rw.lineTail = []byte(lineTail)
	return nil
}

func (rw *BinFileReadWrite) readFull(r io.Reader, buf []byte) error {
	index := 0
	for index < len(buf) {
		n, err := r.Read(buf[index:])
		if err != nil {
			return err
		}
		index += n             //累计读取量
		rw.readOff += int64(n) //累计读偏移
	}
	return nil
}

//打开文件
//fileName  	是输出文件名
//fileSync   	是输入是否同步写文件
//err   	   	是输出错误信息
func (rw *BinFileReadWrite) Open(fileName string, fileSync bool) error {
	flag := os.O_RDWR
	if fileSync {
		flag |= os.O_SYNC
	}
	fd, err := os.OpenFile(fileName, flag, 0660) //建议同步写
	if err != nil {                              //文件不存在
		return err
	}

	rw.lineEof.Reset()

	if rw.fileName != "" && rw.fd != nil {
		rw.fd.Close()
	}

	rw.fileName = fileName
	rw.flock = flock.NewFlock(fileName + fwrite.LockSuffix)
	rw.fd = fd
	rw.readOff = 0
	rw.reader = bufio.NewReader(fd)
	return nil
}

//关闭文件
func (rw *BinFileReadWrite) Close() error {
	if rw.fd == nil || rw.reader == nil {
		return ErrFileNotOpen
	}
	if e := rw.fd.Close(); e != nil {
		return e
	}
	rw.fd = nil
	rw.reader = nil
	rw.lineEof.Reset()
	return nil
}

//读取文件行
//line  	是输入文件行
func (rw *BinFileReadWrite) Read(line *FileLine) error {
	if rw.fd == nil || rw.reader == nil {
		return ErrFileNotOpen
	}

	if rw.lineEof.IsEof {
		line.Clone(rw.lineEof)
		return nil
	}

	line.Reset()
	line.FileName = rw.fileName

	//读取二进制数据头           1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16
	var lineHeadByte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	e := rw.readFull(rw.reader, lineHeadByte) //读取行头
	if e != nil {
		return e
	}

	lineSize := int(binary.BigEndian.Uint32(lineHeadByte[:4]))
	lineVer := lineHeadByte[4]
	if lineVer != 1 {
		return errorf("line version is't 1\n")
	}
	headSize := lineHeadByte[5]
	if headSize != 8 {
		return errorf("line head size is't 8\n")
	}
	markSize := lineHeadByte[6]
	//space := lineHeadByte[7]
	line.LineNO = int64(binary.BigEndian.Uint32(lineHeadByte[8:12]))
	bordSize := int(binary.BigEndian.Uint32(lineHeadByte[12:16]))
	if n := lineSize - FixSize - HeadSize - bordSize - int(markSize); n != 0 {
		return errorf("line size mismatching")
	}
	//读取二进制数据
	line.buff.Reset()
	line.buff.Grow(bordSize)
	buff := line.buff.Bytes()[:bordSize]
	rw.readFull(rw.reader, buff) //读取二进制数据
	if e != nil {
		return e
	}
	line.Line.Write(buff)

	//读取二进制数据标记
	line.buff.Reset()
	line.buff.Grow(int(markSize))
	buff = line.buff.Bytes()[:markSize]
	rw.readFull(rw.reader, buff) //读取二进制数据
	if e != nil {
		return e
	}
	line.Mark = trimMark(string(buff))
	line.use = len(line.Mark)
	line.free = lineSize - FixSize - HeadSize - bordSize - line.use
	line.off = rw.readOff - int64(line.free)
	line.readFD = rw.fd
	line.IsEof = line.LineNO == 0 && line.Mark == fwrite.FileEof
	if line.IsEof {
		rw.lineEof.Clone(line)
	}
	return nil
}

//检查文件是否锁定
func (rw *BinFileReadWrite) Locked() bool {
	if rw.fd == nil {
		panic("文件没有打开")
	}

	if e := rw.flock.NBLock(); e == nil {
		rw.flock.Unlock()
		return false
	} else {
		return true
	}
}

//读取文件行
//line  	是输入文件行
//make  	是输入文件行标记
func (rw *BinFileReadWrite) Mark(line *FileLine, mark string) error {
	if rw.fd == nil || rw.reader == nil {
		return ErrFileNotOpen
	}

	if line.readFD != rw.fd {
		return ErrLineNotMatch
	}

	if line.off <= 0 {
		return errorf("line off less equi than 0")
	}
	if line.use < 0 {
		return errorf("line use less than 0")
	}
	mark = strings.TrimSpace(mark)
	if len(mark) > line.free {
		return errorf("line mark len more than %v", line.free)
	}
	var e error
	if line.use > 0 {
		line.Mark = line.Mark + "," + mark
		_, e = rw.fd.WriteAt([]byte(","+mark), line.off)
	} else {
		line.Mark = mark
		_, e = rw.fd.WriteAt([]byte(mark), line.off)
	}
	return e
}

//释放所有资源
func (rw *BinFileReadWrite) Destroy() {
	if rw.fd != nil {
		rw.fd.Close()
		rw.fd = nil
		rw.reader = nil
	}
}

//写入缓存数据
func (rw *BinFileReadWrite) Flush() {
	if rw.fd != nil {
		rw.fd.Sync()
	}
}
