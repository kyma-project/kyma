package restore

import (
	"bytes"
	"os/exec"
)

func runCommand(command string, args []string) (bytes.Buffer, bytes.Buffer, error) {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout, stderr, err
}
