package process

import (
	"bytes"
	"net"

	"../models"
)

// executer operation
func executerOperation(conn net.Conn, operateData models.OperateModel) {
	var f bytes.Buffer
	// set
	switch {
	case operateData.Operate == "set":
		// write to tikv
		//
		// return for memcache
		conn.Write([]byte("STORED\r\n"))
	case operateData.Operate == "add":
		// write to tikv
		//
		// return for memcache
		conn.Write([]byte("NOT_STORED\r\n"))
	case operateData.Operate == "gets":
		// write to tikv
		//
		// return for memcache
		f.Write([]byte("VALUE foo 0 3 20\r\n"))
		f.Write([]byte("bar\r\n"))
		f.Write([]byte("END\r\n"))
		conn.Write([]byte(f.String()))
	default:
	}
}
