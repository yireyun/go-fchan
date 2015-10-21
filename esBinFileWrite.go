package fchan

import (
	"encoding/binary"
	"fmt"
	"github.com/yireyun/go-fwrite"
	"math"
	"strings"
)

//只写文件记录器
type BinFileWrite struct {
	fwrite.FileWrite

	//行标记长度
	lineMark string

	cfg *FileConfig
}

//创建只写文件记录器
//fileName	是出文件名
//err   	是输出错误信息
func NewBinFileWrite(name string) *BinFileWrite {
	w := new(BinFileWrite)

	w.lineMark = LineMark

	w.cfg = new(FileConfig)
	w.cfg.InitAsDefault(name)
	w.cfg.RotateRenameSuffix = true
	w.cfg.CleanRenameSuffix = true
	w.cfg.FileLock = true
	w.InitFileWriter(name, w.cfg)
	return w
}

//初始化行标志和标记
//markSize	是输入行尾预留尺寸，不能小于32
func (w *BinFileWrite) InitLineMark(markSize int) error {
	if markSize < 32 {
		return fmt.Errorf("line mark len is not less than 32")
	}
	if markSize > 128 {
		return fmt.Errorf("line mark len is not great than 128")
	}
	w.lineMark = strings.Repeat(" ", markSize)
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
		return fmt.Errorf("line mark len more than %v", len(w.lineMark))
	}
	if len(line.Line.Bytes()) > math.MaxInt32 {
		return fmt.Errorf("line bytes len is not great than MaxInt32")
	}
	line.buff.Reset()
	in := line.Line.Bytes()
	//进行文件旋转检查
	line.FileName, line.LineNO = w.cfg.RotateCheck(&w.FileWrite, len(in))
	//写入行头
	var lineHeadByte = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.BigEndian.PutUint32(lineHeadByte[:4], uint32(len(in)+len(w.lineMark)+1))
	binary.BigEndian.PutUint32(lineHeadByte[4:8], uint32(line.LineNO))
	binary.BigEndian.PutUint32(lineHeadByte[8:12], uint32(len(in)))
	line.buff.Write(lineHeadByte)
	//写入行内容
	line.buff.Write(in)
	//写入行尾
	mark := []byte(w.lineMark + "\n")
	line.use = len(line.Mark)
	copy(mark, line.Mark)
	line.buff.Write(mark)
	//最后写入数据
	_, err = w.cfg.MutexWriter(&w.FileWrite, line.buff.Bytes())
	return
}
