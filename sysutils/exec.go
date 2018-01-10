package sysutils

import (
	"bytes"
	"os/exec"
)

//It will return the output for stdout,stderr
func ExecuteCommand(cmd *exec.Cmd) (error, string, string) {
	var procOut bytes.Buffer
	var procErr bytes.Buffer
	cmd.Stdout = &procOut
	cmd.Stderr = &procErr
	err := cmd.Run()
	if err != nil {
		return err, "", procErr.String()
	}
	return nil, procOut.String(), ""
}
