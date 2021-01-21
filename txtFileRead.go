package fchan

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wukehong/lock"
	"github.com/yireyun/go-fwrite"
)

//文件记录器
type TxtFileReadWrite struct {
	Name     string
	fileName string
	flock    lock.Flocker
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
func NewTxtFileReadWrite(name string) *TxtFileReadWrite {
	rw := new(TxtFileReadWrite)
	rw.Name = name
	rw.lineHead = []byte(LineHead)
	rw.lineTail = []byte(LineTail)
	rw.lineMark = LineMark
	return rw
}

//初始化行标志和标记
//lineHead	是输入行头标志，如:"---"，长度不能小于1
//lineTail	是输入行尾标志，如:"==="，长度不能小于1
func (rw *TxtFileReadWrite) InitLineMark(lineHead, lineTail string) error {
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

//打开文件
//fileName  	是输出文件名
//fileSync   	是输入是否同步写文件
//err   	   	是输出错误信息
func (rw *TxtFileReadWrite) Open(fileName string, fileSync bool) error {
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
	rw.flock = lock.NewFlock(fileName + fwrite.LockSuffix)
	rw.fd = fd
	rw.readOff = 0
	rw.reader = bufio.NewReader(fd)
	return nil
}

//关闭文件
func (rw *TxtFileReadWrite) Close() error {
	if rw.fd != nil {
		rw.fd.Close()
		rw.fd = nil
		rw.reader = nil
	}
	return fmt.Errorf("file not open")
}

//读取文件行
//line  	是输入文件行
func (rw *TxtFileReadWrite) Read(line *FileLine) error {
	if rw.fd == nil || rw.reader == nil {
		return fmt.Errorf("file not open")
	}

	var b, head, tail []byte
	var e error

	line.Reset()
	line.FileName = rw.fileName

	for e == nil {
		b, e = rw.reader.ReadBytes('\n') //读取行数据
		rw.readOff += int64(len(b))      //累计行偏移
		if e != nil && len(b) == 0 {
			return e
		}

		//判断是否是行头
		switch {
		case bytes.HasPrefix(b, rw.lineHead): //发现行头
			if len(head) > 0 { //行头重复
				return fmt.Errorf("%s read line head repeat", rw.Name)
			}
			head = make([]byte, len(b))                                  //初始化行头
			copy(head, b)                                                //复制行头
			lineNo := strings.TrimSpace(string(head[len(rw.lineHead):])) //截取行号
			line.LineNO, e = strconv.ParseInt(lineNo, 10, 64)            //解析行号
			if e != nil {                                                //判断解析结果
				return fmt.Errorf("%s read line num error:%v", rw.Name, e)
			}
			tail = make([]byte, 0, len(b)+1+len(rw.lineMark)) //初始化行未
			tail = append(tail, rw.lineTail...)               //设置行尾符号
			tail = append(tail, lineNo...)                    //设置行号
			tail = append(tail, ';')                          //设置分隔符
		case bytes.HasPrefix(b, rw.lineTail): //发现行未
			if line.LineNO <= 0 { //行头缺少
				return fmt.Errorf("%s read line head miss", rw.Name)
			}
			if !bytes.HasPrefix(b, tail) {
				return fmt.Errorf("%s read line num not equi", rw.Name)
			}
			line.Mark = trimMark(string(b[len(tail):]))       //设置行尾
			line.use = len(line.Mark)                         //设置已用
			line.free = len(b[len(tail):len(b)-1]) - line.use //设置空闲
			line.off = rw.readOff - int64(line.free) - 1      //设置偏移
			return nil
		default:
			if line.LineNO <= 0 { //行头缺少
				return fmt.Errorf("%s read line head miss", rw.Name)
			}
			_, e = line.Line.Write(b)
			if e != nil {
				return fmt.Errorf("%s buff line error:%v", rw.Name, e)
			}
		}
	}
	return nil
}

//检查文件是否锁定
func (rw *TxtFileReadWrite) Locked() bool {
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
func (rw *TxtFileReadWrite) Mark(line *FileLine, mark string) error {
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
func (rw *TxtFileReadWrite) Destroy() {
	if rw.fd != nil {
		rw.fd.Close()
		rw.fd = nil
		rw.reader = nil
	}
}

//写入缓存数据
func (rw *TxtFileReadWrite) Flush() {
	if rw.fd != nil {
		rw.fd.Sync()
	}
}
