package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type flags struct {
	count int
}

func main() {
	var args = flags{}
	flag.IntVar(&args.count, "count", 1, "Number of log pipelines to deploy")
	flag.Parse()

	out := os.Stdout
	if err := run(out, args); err != nil {
		fmt.Fprintf(out, "Error: %v\n", err)
		os.Exit(2)
	}

}

func run(out io.Writer, f flags) error {

}
