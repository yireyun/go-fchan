// txtBsonWrite_test
package fchan

import (
	"encoding/base64"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	No   int
	Name string
	Age  int
	Sex  int
}

func TestTxtWrite(t *testing.T) {
	t.SkipNow()
	w := NewTxtFileWrite("FileWriter")
	//fileSync, filePrefix, writeSuffix, renameSuffix string,
	//rotate, dayend,zeroSize bool, maxLines, maxSize int,
	//cleaning bool, maxDays int
	_, err := w.Init(true, "testTxtW", "jour", "jour", "bak",
		true, true, true, 10000*101, 0, false, 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	N := 100

	p := &Person{
		No:   1,
		Name: "hhee",
		Age:  12,
		Sex:  1,
	}
	for i := 1; i < N; i++ {
		p.No = i
		buff, err := bson.Marshal(p)
		if err != nil {
			t.Fatal(err)
		}
		wLine := NewFileLine()
		wLine.Line.WriteString(base64.StdEncoding.EncodeToString(buff))
		e := w.Write(wLine)
		if e != nil {
			t.Fatal(e)
		}
		time.Sleep(time.Second)
		if i%10 == 0 {
			err = w.Rotate()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
