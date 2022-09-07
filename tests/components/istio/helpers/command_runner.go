package helpers

import (
	"bufio"
	"fmt"
	"os/exec"
)

type Command struct {
	Cmd           string
	Args          []string
	OutputChannel chan string
}

func (c *Command) Run() (chan struct{}, error) {
	cmd := exec.Command(c.Cmd, c.Args...)
	r, _ := cmd.StdoutPipe()

	// Use the same pipe for standard error
	cmd.Stderr = cmd.Stdout

	// Make a new channel which will be used to ensure we get all output
	done := make(chan struct{})

	// Create a scanner which scans r in a line-by-line fashion
	scanner := bufio.NewScanner(r)

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {

		// Read line by line and process it
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			c.OutputChannel <- line
		}

		// We're all done, unblock the channel
		done <- struct{}{}

	}()

	// Start the command and check for errors
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	go cmd.Wait()

	return done, nil
}
