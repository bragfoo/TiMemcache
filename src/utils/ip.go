package util

import (
	"net"
	"strings"
)

// LoatlIP is make local ip
var LoatlIP string

// GetIP is get local ip
func GetIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	LoatlIP = strings.Split(addrs[1].String(), "/")[0]
	CustomLogger("local ip:", LoatlIP)
}
