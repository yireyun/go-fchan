package fchan

import (
	"bufio"
	"encoding/binary"
	"fmt"
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
	flock    flock.Flocker
	fd       *os.File
	reader   *bufio.Reader
	readOff  int64
	lineHead []byte
	lineTail []byte
	lineMark string
}

//创建文件记录器
//fileName	是出文件名
//err   	是输出错误信息
func NewBinFileReadWrite(name string) *BinFileReadWrite {
	rw := new(BinFileReadWrite)
	rw.Name = name
	rw.lineHead = []byte(LineHead)
	rw.lineTail = []byte(LineTail)
	rw.lineMark = LineMark
	return rw
}

//初始化行标志和标记
//lineHead	是输入行头标志，如:"---"，长度不能小于1
//lineTail	是输入行尾标志，如:"==="，长度不能小于1
func (rw *BinFileReadWrite) InitLineMark(lineHead, lineTail string) error {
	lineHead = strings.TrimSpace(lineHead)
	lineTail = strings.TrimSpace(lineTail)

	if len(lineHead) < 0 {
		return fmt.Errorf("line head is null")
	}
	if len(lineTail) < 0 {
		return fmt.Errorf("line Tail is null")
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
	if rw.fd != nil {
		rw.fd.Close()
		rw.fd = nil
		rw.reader = nil
	}
	return fmt.Errorf("file not open")
}

//读取文件行
//line  	是输入文件行
func (rw *BinFileReadWrite) Read(line *FileLine) error {
	if rw.fd == nil || rw.reader == nil {
		return fmt.Errorf("file not open")
	}

	line.Reset()
	line.FileName = rw.fileName

	//读取二进制数据头
	var lineHeadByte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	e := rw.readFull(rw.reader, lineHeadByte) //读取行头
	if e != nil {
		return e
	}
	totalLen := int(binary.BigEndian.Uint32(lineHeadByte[:4]))
	line.LineNO = int64(binary.BigEndian.Uint32(lineHeadByte[4:8]))
	bordLen := int(binary.BigEndian.Uint32(lineHeadByte[8:12]))
	if n := totalLen - len(lineHeadByte); bordLen >= n {
		return fmt.Errorf("bord size is greater than %d", n)
	}
	//读取二进制数据
	line.buff.Reset()
	line.buff.Grow(bordLen)
	buff := line.buff.Bytes()[:bordLen]
	rw.readFull(rw.reader, buff) //读取二进制数据
	if e != nil {
		return e
	}
	line.Line.Write(buff)

	//读取二进制数据标记
	line.buff.Reset()
	line.buff.Grow(totalLen - bordLen)
	buff = line.buff.Bytes()[:totalLen-bordLen]
	rw.readFull(rw.reader, buff) //读取二进制数据
	if e != nil {
		return e
	}
	line.Mark = trimMark(string(buff))
	line.use = len(line.Mark)
	line.free = totalLen - bordLen - line.use - 1
	line.off = rw.readOff - int64(line.free) - 1
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
		return fmt.Errorf("file not open")
	}

	if line.off <= 0 {
		return fmt.Errorf("line off less equi than 0")
	}
	if line.use < 0 {
		return fmt.Errorf("line use less than 0")
	}
	mark = strings.TrimSpace(mark)
	if len(mark) > line.free {
		return fmt.Errorf("line mark len more than %v", line.free)
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
