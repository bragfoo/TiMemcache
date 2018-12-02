package process

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"../models"
)

// dispose request
func disposeRequest(conn net.Conn) {
	var textData models.OperateModel
	var f bytes.Buffer
	buf := make([]byte, 512)
	ip := conn.RemoteAddr().String()
	// close connection before exit
	defer func() {
		log.Println("Disconnected with " + ip)
		conn.Close()
	}()
	// for connect reuse & keep-live
	for {
		// read from tcp
		_, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Client %s is close", ip)
			}
			break
		}
		log.Println("Line info:", string(buf))
		lineSlices := strings.Split(string(buf), "\r\n")
		// text line
		textLine := lineSlices[0]
		// analysis operation
		textSlices := strings.Split(textLine, " ")
		textData = analysisOperation(textSlices)
		// free line
		for n, line := range lineSlices {
			if n != 0 {
				// write to buffer
				f.Write([]byte(line))
			}
		}
		textData.Value = []byte(f.String())
		// executer operation
		executerOperation(conn, textData)
	}
}

// analysis operation
func analysisOperation(cmdList []string) models.OperateModel {
	textData := models.OperateModel{}
	cmdOperation := cmdList[0]
	key := cmdList[1]
	// set
	if cmdOperation == "set" {
		textData.Operate = "set"
		textData.Key = key
		// add
	} else if cmdOperation == "add" {
		textData.Operate = "add"
		textData.Key = key
		// delete
	} else if cmdOperation == "delete" {
		// log.Println(cmdOperation, key)
	} else if cmdOperation == "replace" {
		// gets
	} else if cmdOperation == "gets" {
		textData.Operate = "gets"
		textData.Key = key
	} else {
		log.Println("Operation invalid")
	}
	return textData
}
