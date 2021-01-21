package fchan

import (
	"fmt"
	"io"
	"testing"
	"time"
)

func TestTxtReadWrite(t *testing.T) {
	//t.SkipNow()
	w := NewTxtFileWrite("TestWrite")
	//fileSync, filePrefix, writeSuffix, renameSuffix, cleanSuffix,
	//rotate, dayend, zeroSize, maxLines, maxSize, clean, maxDays, lastFiler
	filename, err := w.Init(true, "testTxtRW", "jour", "jour", "bak",
		true, true, true, 10000*10+5, 0, false, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := NewTxtFileReadWrite("TxtFileReadWrite")
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

	isPrint := false
	lineNO := int64(0)
	writeReadWrite := func() {
		e := w.Write(wLine)
		if e != nil {
			t.Fatal(e)
		} else {
			lineNO++
		}
		if isPrint {
			fmt.Printf("Write:%+v\n", wLine)
		}
		e = rw.Read(rLine)
		if e != nil {
			t.Fatal(e)
		}
		if lineNO != rLine.LineNO {
			t.Fatalf("lineNO: expect %d is not %d\n%v\n",
				lineNO, rLine.LineNO, rLine.String())
		}
		if isPrint {
			fmt.Printf("Read :%+v\n", rLine)
		}
		e = rw.Mark(rLine, "Ok")
		if e != nil {
			t.Fatal(e)
		}
		if isPrint {
			fmt.Printf("Mark :%+v\n", rLine)
		}
	}

	start := time.Now()

	writeReadWrite()

	writeReadWrite()

	writeReadWrite()

	writeReadWrite()

	writeReadWrite()

	N := 100 * 10
	go func() {
		defer func() {
			w.Close()
		}()
		for i := 0; i < N; i++ {
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
		for record <= N { //多读尾行EOF
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

			if rLine.IsEof {
				t.Logf("Read '%s' EOF\n", rLine.FileName)
				return
			}

			if lineNO != rLine.LineNO {
				t.Fatalf("lineNO: expect %d is not %d\n%v\n",
					lineNO, rLine.LineNO, rLine.String())
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
