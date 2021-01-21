package fchan

import (
	"testing"
	"time"
)

func TestTxtWriteFile(t *testing.T) {
	t.SkipNow()
	w := NewTxtFileWrite("Test")
	var err error
	//fileSync, filePrefix, writeSuffix, renameSuffix, cleanSuffix,
	//rotate, dayend, zeroSize, maxLines, maxSize, clean, maxDays, lastFiler
	_, err = w.Init(true, "TestTxtWt", "jour", "jour", "bak",
		true, true, true, 10, 1<<20, true, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("012345678901234567890123456789012345678901234567890123456789")
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	//	msg = append(msg, []byte("012345678901234567890123456789012345678901234567890123456789")...)
	line := NewFileLine()
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
	if e := w.Close(); e != nil {
		t.Errorf("Close: %v\n", e)
	} else {
		t.Logf("Close: Succ\n")
	}
}
