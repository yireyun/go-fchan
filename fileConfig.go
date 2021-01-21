package fchan

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/yireyun/go-fwrite"
)

type FileConfig struct {
	fwrite.FileConfig
}

func (c *FileConfig) Config() *fwrite.FileConfig {
	return &c.FileConfig
}

func (w *FileConfig) GetFileRename(fileName string) (fileRename string, err error) {
	if fileName == "" {
		return "", fmt.Errorf("get file rename, fileName is null")
	}
	if w.RenameSuffix == "" {
		return "", fmt.Errorf("get file name, renameSuffix is null")
	}
	if w.WriteSuffix == "" {
		return "", fmt.Errorf("get file name, rriteSuffix is null")
	}

	if w.RotateRenameSuffix && w.RenameSuffix == w.WriteSuffix &&
		strings.HasSuffix(fileName, w.RenameSuffix) {
		return fileName, nil
	}

	if strings.HasSuffix(fileName, w.WriteSuffix) {
		fileName = fileName[:len(fileName)-len(w.WriteSuffix)]
	} else {
		return fileName, fmt.Errorf("get file rename, suffix is error")
	}

	fileRename = fileName + w.RenameSuffix
	if _, err = os.Lstat(fileRename); err == nil {
		for num := 1; err == nil && num <= math.MaxInt16; num++ { //出现重名时增加序号
			fileRename = fmt.Sprintf("%s.%03d%s", fileName, num, w.RenameSuffix)
			_, err = os.Lstat(fileRename)
		}
	}
	if err == nil {
		err = fmt.Errorf("Cannot find free file full number:%s", fileName)
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
		return "", fmt.Errorf("get file name, filePrefix is null")
	}
	if w.WriteSuffix == "" {
		return "", fmt.Errorf("get file name, writeSuffix is null")
	}
	if w.RenameSuffix == "" {
		return "", fmt.Errorf("get file name, renameSuffix is null")
	}
	now := time.Now()
	for num := 1; num <= math.MaxInt16; num++ {
		curDate := now.Format("2006-01-02")
		fileName := fmt.Sprintf("%s.%s.%03d%s", w.FilePrefix, curDate, num, w.WriteSuffix)
		fileRename := fmt.Sprintf("%s.%s.%03d%s", w.FilePrefix, curDate, num, w.RenameSuffix)
		fileClean := fmt.Sprintf("%s.%s.%03d%s", w.FilePrefix, curDate, num, w.CleanSuffix)
		_, fileNameErr := os.Lstat(fileName)
		_, fileRenameErr := os.Lstat(fileRename)
		_, fileCleanErr := os.Lstat(fileClean)
		if w.RotateRename && fileNameErr == nil && w.FileName != fileName { //文件重命名
			newName, err := w.GetFileRename(fileName)
			if err == nil {
				err = os.Rename(fileName, newName)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Printf("\t%s rename '%s' error:%v\n", w.Name, fileName, err)
			}
		}

		if fileNameErr != nil && fileRenameErr != nil && fileCleanErr != nil {
			return fileName, nil
		}
	}
	return "", fmt.Errorf("Cannot find free file name number:%s", w.FilePrefix)
}
