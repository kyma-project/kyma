package sh

import "os/exec"

func RunInDir(command, dir string) (string, error) {
	cmd := execCommand(command)
	cmd.Dir = dir
	return strOutput(cmd)
}

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
