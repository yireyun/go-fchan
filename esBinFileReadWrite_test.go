package fchan

import (
	"fmt"
	"io"
	"testing"
	"time"
)

func TestBinReadWrite(t *testing.T) {
	//	t.SkipNow()
	w := NewBinFileWrite("FileWrite")
	//fileSync, filePrefix, writeSuffix, renameSuffix string,
	//rotate, dayend bool, maxLines, maxSize int,
	//cleaning bool, maxDays int
	filename, err := w.Init(true, "testBinRW", "log", "log", "log", true, true,
		10000*10+5, 0, false, 3)
	if err != nil {
		t.Fatal(err)
	}
	rw := NewBinFileReadWrite("BinFileReadWrite")
	err = rw.Open(filename, true)
	if err != nil {
		t.Fatal(err)
	}

	wLine := NewFileLine()
	wLine.Mark = "mark"
	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")
	//	wLine.Line.WriteString("\t文件日志，文件日志，文件日志，文件日志，文件日志，文件日志。\n")

	rLine := NewFileLine()

	print := false
	lineNO := int64(0)
	writeReadWrite := func() {
		e := w.Write(wLine)
		if e != nil {
			t.Fatal(e)
		} else {
			lineNO++
		}
		if print {
			fmt.Printf("Write:%+v\n", wLine)
		}
		e = rw.Read(rLine)
		if e != nil {
			t.Fatal(e)
		}
		if lineNO != rLine.LineNO {
			t.Fatalf("lineNO: expect %d is not %d", lineNO, rLine.LineNO)
		}
		if print {
			fmt.Printf("Read :%+v\n", rLine)
		}
		e = rw.Mark(rLine, "Ok")
		if e != nil {
			t.Fatal(e)
		}
		if print {
			fmt.Printf("Mark :%+v\n", rLine)
		}
	}

	start := time.Now()

	writeReadWrite()

	writeReadWrite()

	writeReadWrite()

	writeReadWrite()

	writeReadWrite()

	N := 10000 * 10
	writeStop := false
	go func() {
		defer func() { writeStop = true }()
		for i := 0; i <= N; i++ {
			e := w.Write(wLine)
			if e != nil {
				t.Fatal(e)
			}
		}
	}()
	readStopC := make(chan int, 1)
	go func() {
		defer func() { readStopC <- 0 }()

		record := 0
		for record < N {
			e := rw.Read(rLine)
			if e == io.EOF {
				if rw.Locked() {
					continue
				} else {
					t.Error("file is unlock")
					return
				}
			}

			if e != nil {
				t.Fatal(e)
			} else {
				lineNO++
				record++
			}
			if lineNO != rLine.LineNO {
				fmt.Printf("Read :%+v\n", rLine)
				t.Fatalf("lineNO: expect %d is not %d", lineNO, rLine.LineNO)
			}

			if record%10 == 0 {
				e = rw.Mark(rLine, "mongdb=err")
			} else {
				e = rw.Mark(rLine, "mongdb=ok")
			}
			if e != nil {
				t.Fatal(e)
			}
		}
	}()
	<-readStopC
	end := time.Now()
	t.Logf("CNT:%d,Use:%s\n", N+5, end.Sub(start))
	if !rw.Locked() {
		t.Logf("file '%s' is  unlock", filename)
	}
}
