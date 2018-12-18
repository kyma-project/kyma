package sh

import "os/exec"

// RunInDir runs given command in shell in specific directory
func RunInDir(command, dir string) (string, error) {
	cmd := execCommand(command)
	cmd.Dir = dir
	return strOutput(cmd)
}

// Run runs given command in shell
func Run(command string) (string, error) {
	cmd := execCommand(command)
	return strOutput(cmd)
}

func execCommand(command string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", command)
}

func strOutput(cmd *exec.Cmd) (string, error) {
	out, err := cmd.CombinedOutput()
	return string(out), err
}
