package test

import (
	stdlog "log"
	"os"
)

// FileExists checks if specified file exists.
func FileExists(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		stdlog.Fatalf("File %s does not exists", filename)
	} else if err != nil {
		stdlog.Fatal(err)
	}
}
