package fchan

import (
	"math"
	"os"
	"strings"
	"time"

	"github.com/yireyun/go-fwrite"
)

type LastFiler interface {
	LastFile() *fileInfo
}

type FileConfig struct {
	fwrite.FileConfig
	lastFile LastFiler
}

func NewFileConfig() *FileConfig {
	return &FileConfig{}
}

func (c *FileConfig) Config() *fwrite.FileConfig {
	return &c.FileConfig
}

func (w *FileConfig) GetFileRename(fileName string) (fileRename string, err error) {
	if fileName == "" {
		return "", errorf("get file rename, fileName is null")
	}
	if w.RenameSuffix == "" {
		return "", errorf("get file name, renameSuffix is null")
	}
	if w.WriteSuffix == "" {
		return "", errorf("get file name, rriteSuffix is null")
	}

	if w.RotateRenameSuffix && w.RenameSuffix == w.WriteSuffix &&
		strings.HasSuffix(fileName, w.RenameSuffix) {
		return fileName, nil
	}

	if strings.HasSuffix(fileName, w.WriteSuffix) {
		fileName = fileName[:len(fileName)-len(w.WriteSuffix)]
	} else {
		return "", errorf("get file rename, suffix is error")
	}

	fileRename = fileName + w.RenameSuffix
	exist := fwrite.FileExist(fileRename)
	if exist {
		for num := 1; exist && num <= math.MaxInt16; num++ {
			fileRename = sprintf("%s.%03d%s", fileName, num, w.RenameSuffix)
			exist = fwrite.FileExist(fileRename)
		}
	}
	if exist {
		err = errorf("Cannot find free file full number:%s", fileName)
	} else {
		err = nil
	}
	return
}

//获取文件行数
//fileName	是输入文件名
func (c *FileConfig) GetFileLines(fileName string) (int64, error) {
	return c.FileConfig.GetFileLines(fileName)
}

//获取重命名文件名
//fileName  	是输入文件名
//fileRename	是输出重命名文件名
//err       	是输出错误信息
func (w *FileConfig) GetNewFileName() (string, error) {
	if w.FilePrefix == "" {
		return "", errorf("get file name, filePrefix is null")
	}
	if w.WriteSuffix == "" {
		return "", errorf("get file name, writeSuffix is null")
	}
	if w.RenameSuffix == "" {
		return "", errorf("get file name, renameSuffix is null")
	}
	now := time.Now()
	var fileDate = ""
	lastTime, lastNum, ok := w.getLastTime()
	if ok {
		fileDate, lastNum = lastTime.Format("2006-01-02"), lastNum+1
		if nowDate := now.Format("2006-01-02"); fileDate >= nowDate {
			printf("<TRACE> GetNewFileName SUCC : %v, %03v\n\n", fileDate, lastNum)
		} else {
			fileDate, lastNum = nowDate, 1
		}
	} else {
		fileDate, lastNum = now.Format("2006-01-02"), 1
		if w.lastFile != nil {
			printf("<ERROR> GetNewFileName FAIL : %v, %03v\n\n", fileDate, lastNum)
		}
	}

	for num := lastNum; num <= math.MaxInt16; num++ {
		fileName := sprintf("%s.%s.%03d%s", w.FilePrefix, fileDate, num, w.WriteSuffix)
		fileRename := sprintf("%s.%s.%03d%s", w.FilePrefix, fileDate, num, w.RenameSuffix)
		fileClean := sprintf("%s.%s.%03d%s", w.FilePrefix, fileDate, num, w.CleanSuffix)
		fileNameExist := fwrite.FileExist(fileName)
		fileRenameExist := fwrite.FileExist(fileRename)
		fileCleanExist := fwrite.FileExist(fileClean)
		fileLocked := fwrite.FileLocked(fileName)
		if w.RotateRename && fileNameExist && !fileLocked &&
			w.FileName != fileName { //文件重命名
			newName, err := w.GetFileRename(fileName)
			if err == nil {
				err = os.Rename(fileName, newName)
				if err != nil {
					printf("<ERROR> '%s' GetNewFileName() os.Rename \"%s\" -> \"%s\" Error: %v\n",
						w.Name, fileName, newName, err)
				}
			} else {
				printf("<ERROR> '%s' GetNewFileName() \"%s\" Error: %v\n", w.Name, fileName, err)
			}
		}

		if !fileNameExist && !fileRenameExist && !fileCleanExist {
			return fileName, nil
		}
	}
	return "", errorf("Cannot find free file name number:%s", w.FilePrefix)
}

//获取上次时间和序号, 如读取不到, return time.Time{},0
//fileName	是输入文件名
//lastTime	是输出上次的时间
//lastNum	是输出上次的序号
//ok    	是输出操作结果
func (c *FileConfig) getLastTime() (lastTime time.Time, lastNum int, ok bool) {
	if c == nil || c.lastFile == nil {
		return time.Time{}, 0, false
	}
	if last := c.lastFile.LastFile(); last != nil {
		return last.fileTime, last.fileIndex, true
	} else {
		return time.Time{}, 0, false
	}
}
