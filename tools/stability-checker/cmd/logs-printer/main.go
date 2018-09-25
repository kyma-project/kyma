package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	podlogger "github.com/kyma-project/kyma/tools/stability-checker/internal/log"
	"github.com/kyma-project/kyma/tools/stability-checker/internal/printer"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config holds configuration for logs-printer application
type Config struct {
	WorkingNamespace string
	PodName          string `envconfig:"HOSTNAME"`
}

func main() {
	var requestedTestIDs ids
	flag.Var(&requestedTestIDs, "ids", "A comma separated list of tests ids.")
	flag.Parse()

	if requestedTestIDs == nil {
		log.Printf("Parameter --ids was not provided. All failed tests outputs will be printed.")
	}

	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(errors.Wrap(err, "while reading configuration from environment variables"))

	logFetcher, err := podlogger.NewPodLogFetcher(cfg.WorkingNamespace, cfg.PodName)
	fatalOnError(err)
	logReadCloser, err := logFetcher.GetLogsFromPod()
	fatalOnError(errors.Wrap(err, "while getting logs from pod"))
	defer logReadCloser.Close()

	stream := json.NewDecoder(logReadCloser)

	err = printer.
		New(stream, requestedTestIDs.ToSlice()).
		PrintFailedTestOutput()
	fatalOnError(errors.Wrap(err, "while printing failed tests outputs"))
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

type ids []string

func (s *ids) String() string { return fmt.Sprint(*s) }

func (s *ids) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

func (s *ids) ToSlice() []string {
	return []string(*s)
}
