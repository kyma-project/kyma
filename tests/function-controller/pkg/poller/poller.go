package poller

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Poller struct {
	MaxPollingTime     time.Duration
	InsecureSkipVerify bool
	Log                *logrus.Entry
	DataKey            string
}

func (p Poller) WithLogger(l *logrus.Entry) Poller {
	p.Log = l
	return p
}

func (p Poller) PollForAnswer(url, payloadStr, expected string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: p.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	done := make(chan struct{})

	go func() {
		time.Sleep(p.MaxPollingTime)
		close(done)
	}()

	return wait.PollImmediateUntil(10*time.Second,
		func() (done bool, err error) {
			payload := strings.NewReader(fmt.Sprintf(`{ "%s": "%s" }`, p.DataKey, payloadStr))
			req, err := http.NewRequest(http.MethodGet, url, payload)
			if err != nil {
				return false, errors.Wrapf(err, "while creating new request to ping url %s with payload %s", url, payloadStr)
			}

			req.Header.Add("content-type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				return false, err
			}
			defer func() {
				errClose := res.Body.Close()
				if errClose != nil {
					p.Log.Infof("Error closing body in request to %s: %v", url, errClose)
				}
			}()

			if res.StatusCode != http.StatusOK {
				p.Log.Infof("Expected status %s, got %s, retrying...", http.StatusText(http.StatusOK), res.Status)
				return false, nil
			}

			byteRes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return false, errors.Wrap(err, "while reading response")
			}

			body := string(byteRes)

			if body != expected {
				p.Log.Infof("Got: %q, expected: %s, retrying...", body, expected)
				return false, nil
			}

			p.Log.Infof("Got: %q, correct...", body)
			return true, nil
		}, done)
}
