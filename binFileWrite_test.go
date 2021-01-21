package fchan_test

import (
	"testing"
	"time"

	"github.com/yireyun/go-fchan"
)

func TestBinWriteFile(t *testing.T) {
	t.SkipNow()
	w := fchan.NewBinFileWrite("Test")
	var err error
	//fileSync, filePrefix, writeSuffix, renameSuffix, cleanSuffix,
	//rotate, dayend, fileZip, zeroSize, maxLines, maxSize, clean, maxDays
	_, err = w.Init(true, "TestBinWt", "jour", "jour", "bak",
		true, true, false, true, 100, 1<<20, true, 3)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("012345678901234567890123456789012345678901234567890123456789\n")
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	line := fchan.NewFileLine()
	line.Line.Write(msg)
	line.Mark = "mark"
	start := time.Now()
	var oldName, newName string
	for i := 1; i <= 2*100; i++ {
		err = w.Write(line)
		if err != nil {
			t.Fatal(err)
		}
		newName = line.FileName
		if newName != "" && newName != oldName {
			t.Logf("%05d,logFile:[%s]->[%s],%d", i, oldName, newName, line.LineNO)
			oldName = newName
		}
	}
	end := time.Now()
	if d := end.Sub(start); d < time.Second*10 {
		time.Sleep(d)
	}
}
