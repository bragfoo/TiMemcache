package util

import (
	"os/exec"
)

// RunCmd is run simple cmd
func RunCmd(cmdInfo string) {
	cmd := exec.Command("/bin/bash", "-c", cmdInfo)
	err := cmd.Run()
	if err != nil {
		ErrorLogger(err)
	}
}
