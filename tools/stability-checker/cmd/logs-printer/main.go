package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/kyma-project/kyma/tools/stability-checker/cmd/logs-printer/internal/printer"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
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

	logReader, err := getLogsFromPod(cfg)
	fatalOnError(errors.Wrap(err, "while getting logs from pod"))
	defer logReader.Close()

	stream := json.NewDecoder(logReader)

	err = printer.
		New(stream, requestedTestIDs.ToSlice()).
		PrintFailedTestOutput()
	fatalOnError(errors.Wrap(err, "while printing failed tests outputs"))
}

func getLogsFromPod(cfg Config) (io.ReadCloser, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s config")
	}

	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s client")
	}

	req := k8sCli.CoreV1().Pods(cfg.WorkingNamespace).GetLogs(cfg.PodName, &v1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		return nil, errors.Wrapf(err, "while streaming logs from pod %q", cfg.PodName)
	}

	return readCloser, nil
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
