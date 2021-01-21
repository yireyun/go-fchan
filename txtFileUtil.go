package fchan

import (
	"github.com/yireyun/go-fwrite"
)

var (
	TxtLineHead   = "---"
	TxtLineTail   = "==="
	TxtFmtEof     = "%s%d\n%s%d;%v\n"
	TxtFmtEofHead = "%s%d\n"
	TxtFmtEofTail = "%s%d;%v\n"
)

func GetTxtLineEof(lineHead, lineTail string) (lineEof []byte, tailEof []byte) {
	lineEof = []byte(sprintf(TxtFmtEof, lineHead, 0, lineTail, 0, fwrite.FileEof))
	tailEof = []byte(sprintf(TxtFmtEofTail, lineTail, 0, fwrite.FileEof))
	return
}
