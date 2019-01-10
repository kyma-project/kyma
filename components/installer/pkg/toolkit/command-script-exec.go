package toolkit

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

// CommandExecutor .
type CommandExecutor interface {
	// RunCommand .
	RunCommand(execPath string, execArgs ...string) error
}

//EmptyArgs .
const EmptyArgs string = ""

//KymaCommandExecutor .
type KymaCommandExecutor struct {
}

//RunCommand .
func (kymaBashExecutor *KymaCommandExecutor) RunCommand(execPath string, execArgs ...string) error {
	cmd := exec.Command(execPath, execArgs...)

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()

	if err != nil {
		errStr := string(stderr.Bytes())
		outStr := string(stdout.Bytes())
		log.Println("Bash script error")
		log.Println(strings.Trim(errStr, "\n"))
		log.Println("Bash script output")
		log.Println(strings.Trim(outStr, "\n"))
		return err
	}

	return nil
}
