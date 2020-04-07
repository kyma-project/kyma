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
	legacy  string
	payload string
}

func readFlags() cfg {
	return cfg{
		legacy:  os.Getenv(testsuite.LegacyEnvKey),
		payload: os.Getenv(testsuite.ExpectedPayloadEnvKey),
	}
}

func main() {
	config := readFlags()
	log := logrus.New()
	log.Infof("Legacy: %s, Payload: %s", config.legacy, config.payload)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		reqBytes := make([]byte, 0)
		_, err := req.Body.Read(reqBytes)
		if err != nil {
			res.WriteHeader(403)
			res.Write([]byte("Can't resolve request body"))
			return
		}

		reqBody := fmt.Sprint(reqBytes)
		log.Infof("Received request: %s", reqBody)
		if reqBody != config.payload {
			res.WriteHeader(403)
			res.Write([]byte("Payload not as expected"))
			return
		}

		var gatewayURL string
		if config.legacy == "true" {
			gatewayURL = getLegacyGateway()
		} else {
			gatewayURL = getGateway()
		}

		counterReq := bytes.NewReader([]byte(`{ json: true }`))
		url := fmt.Sprintf("%s/counter", gatewayURL)
		log.Infof("Send %s to %s", counterReq, url)

		postRes, err := http.Post(url, "application/json", counterReq)
		if err != nil {
			log.Infof("Rejected: %s", err)
			return
		}

		resBody := make([]byte, 0)
		_, err = postRes.Body.Read(resBody)
		if err != nil {
			log.Info(err)
			return
		}
		log.Infof("Resolved: %s", resBody)

		res.WriteHeader(200)
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error(err)
	}
}

func getGateway() string {
	for _, env := range os.Environ() {
		keyValue := strings.Split(env, "=")
		if strings.HasSuffix(keyValue[0], "_GATEWAY_URL") {
			return keyValue[1]
		}
	}
	return ""
}

func getLegacyGateway() string {
	return os.Getenv("GATEWAY_URL")
}
