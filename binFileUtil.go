package fchan

import (
	"encoding/binary"

	"github.com/yireyun/go-fwrite"
)

var (
	binTailEof = fwrite.FileEof + "\n"
)

func GetBinLineEof() (lineEof []byte, tailEof []byte) {
	lineEof = append(make([]byte, HeadBytes, HeadBytes), binTailEof...)
	var markSize byte = byte(len(binTailEof))
	var lineSize uint32 = FixSize + HeadSize + uint32(markSize)
	binary.BigEndian.PutUint32(lineEof[:4], lineSize)     //行总长度
	lineEof[4] = FixByte1Ver                              //格式版本
	lineEof[5] = FixByte2HeadSize                         //HAED数据区长度
	lineEof[6] = markSize                                 //MARK标志区长度
	lineEof[7] = 0                                        //空闲
	binary.BigEndian.PutUint32(lineEof[8:12], uint32(0))  //行序号
	binary.BigEndian.PutUint32(lineEof[12:16], uint32(0)) //内容长度
	return
}
