package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	var (
		path       string
		host       string
		statusPort int
		retries    int
	)

	flag.StringVar(&path, "path", "", "liveness or readiness endpoint")
	flag.StringVar(&host, "host", "localhost", "host of the tested application")
	flag.IntVar(&retries, "retry", 1, "number of retries when calling a given endpoint")
	flag.IntVar(&statusPort, "statusPort", 0, "liveness or readiness port")
	flag.Parse()

	url := fmt.Sprintf("http://%s:%d/%s", host, statusPort, path)

	for i := 0; i < retries; i++ {
		resp, err := http.Get(url)
		if err != nil {
			logger.Errorf("while getting response from url: %s: %v", url, err)
			time.Sleep(time.Second)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			os.Exit(0)
		}
		logger.Warnf("%s responded with code %d: Command line arguments : %v", url, resp.StatusCode, os.Args)
		time.Sleep(time.Second)
	}

	os.Exit(1)
}
