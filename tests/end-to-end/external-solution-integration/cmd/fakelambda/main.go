package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	"github.com/sirupsen/logrus"
)

type cfg struct {
	payload string
	legacy  string
}

func readFlags() cfg {
	return cfg{
		payload: os.Getenv(testsuite.ExpectedPayloadEnvKey),
		legacy:  os.Getenv(testsuite.LegacyEnvKey),
	}
}

func main() {
	config := readFlags()
	log := logrus.New()
	log.Infof("Legacy: %s, Payload: %s", config.legacy, config.payload)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(req.Body)
		reqString := strings.Trim(buffer.String(), "\" ")

		log.Infof("Received request: %s", reqString)
		if reqString != config.payload {
			log.Infof("Bad request: %s. Expected %s or \"%s\"", reqString, config.payload, config.payload)
			return
		}

		gateway := getGateway(config.legacy)
		url := fmt.Sprintf("%s/counter", gateway)
		counterReq := bytes.NewReader([]byte(`{ json: true }`))

		log.Infof("Send %s to %s", counterReq., url)
		postRes, err := http.Post(url, "application/json", counterReq)
		if err != nil {
			log.Infof("Rejected: %s", err)
			return
		}

		buffer = new(bytes.Buffer)
		buffer.ReadFrom(postRes.Body)
		resString := buffer.String()
		log.Infof("Resolved: %s", resString)
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error(err)
	}
}

func getGateway(legacy string) (gateway string) {
	if legacy == "true" {
		gateway = os.Getenv("GATEWAY_URL")
	} else {
		for _, env := range os.Environ() {
			keyValue := strings.Split(env, "=")
			if strings.HasSuffix(keyValue[0], "_GATEWAY_URL") {
				gateway = keyValue[1]
			}
		}
	}
	return gateway
}
