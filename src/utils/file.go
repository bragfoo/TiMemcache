package util

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// OverlapWriteFile is overlap write data to file once
func OverlapWriteFile(fileName, fileData string) {
	dirPathSlice := strings.Split(fileName, "/")
	os.MkdirAll(strings.Trim(fileName, dirPathSlice[len(dirPathSlice)-1]), 0755)
	f, fErr := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	if fErr != nil {
		ErrorLogger(fErr)
	}
	_, wErr := f.WriteString(fileData)
	if wErr != nil {
		ErrorLogger(wErr)
	}
	f.Sync()
	f.Close()
}

// MergeFile is merge file
func MergeFile(oldFile, newFile string) {
	fo, foErr := os.OpenFile(oldFile, os.O_RDWR|os.O_APPEND, 0777)
	if foErr != nil {
		ErrorLogger(foErr)
	}
	// open new file
	fn, fnErr := os.Open(newFile)
	if fnErr != nil {
		ErrorLogger(fnErr)
	}
	// read new file
	rd := bufio.NewReader(fn)
	for {
		oneDoc, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		CustomLogger("doc info:", oneDoc)
		// write old file
		_, fwErr := fo.WriteString(oneDoc)
		if fwErr != nil {
			ErrorLogger(fwErr)
		}
	}
	// clone new file
	fn.Close()
	// sync old file
	fo.Sync()
	// close old file
	fo.Close()
}

// DeleteFile is delete file
func DeleteFile(fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		ErrorLogger(err)
	}
}

// DeleteDir is delete dir
func DeleteDir(dirPath string) {
	err := os.RemoveAll(dirPath)
	if err != nil {
		ErrorLogger(err)
	}
}
