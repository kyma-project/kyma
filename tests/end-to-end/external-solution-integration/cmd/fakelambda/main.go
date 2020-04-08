package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	"github.com/sirupsen/logrus"
)

type cfg struct {
	payload   string
	targetUrl string
}

func readFlags() cfg {
	return cfg{
		payload:   os.Getenv(testsuite.ExpectedPayloadEnvKey),
		targetUrl: os.Getenv(testsuite.TargetServiceURLEnvKey),
	}
}

func main() {
	config := readFlags()
	log := logrus.New()
	log.Infof("Target Url: %s, Payload: %s", config.targetUrl, config.payload)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		buffer := new(bytes.Buffer)
		buffer.ReadFrom(req.Body)
		reqString := buffer.String()
		log.Infof("Received request: %s", reqString)
		if reqString != config.payload {
			res.WriteHeader(403)
			res.Write([]byte("Payload not as expected"))
			log.Infof("Bad request: %s expected %s", reqString, config.payload)
			return
		}

		counterReq := bytes.NewReader([]byte(`{ json: true }`))
		url := fmt.Sprintf("%s/counter", config.targetUrl)

		log.Infof("Send %s to %s", counterReq, url)
		postRes, err := http.Post(url, "application/json", counterReq)
		if err != nil {
			log.Infof("Rejected: %s", err)
			res.WriteHeader(403)
			res.Write([]byte("Bad connection with counter"))
			return
		}

		buffer = new(bytes.Buffer)
		buffer.ReadFrom(postRes.Body)
		resString := buffer.String()
		log.Infof("Resolved: %s", resString)

		res.WriteHeader(200)
		res.Write(buffer.Bytes())
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error(err)
	}
}
