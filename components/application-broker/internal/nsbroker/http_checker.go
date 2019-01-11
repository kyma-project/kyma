package nsbroker

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type httpChecker struct {
	log logrus.FieldLogger

	afterFunc func(d time.Duration) <-chan time.Time
	sleepTime time.Duration
}

func newHTTPChecker(log logrus.FieldLogger) *httpChecker {
	return &httpChecker{
		log:       log.WithField("service", "http-checker"),
		afterFunc: time.After,
		sleepTime: time.Second,
	}
}

func (c *httpChecker) WaitUntilIsAvailable(url string, timeout time.Duration) {
	timeoutCh := time.After(timeout)
	for {
		r, err := http.Get(url)
		if err == nil {
			// no need to read the response
			ioutil.ReadAll(r.Body)
			r.Body.Close()
		}
		if err == nil && r.StatusCode == http.StatusOK {
			break
		}

		select {
		case <-timeoutCh:
			c.log.Warnf("Waiting for service %s to be ready timeout %s exceeded.", url, timeout.String())
			if err != nil {
				c.log.Warnf("Last call error: %s", err.Error())
			} else {
				c.log.Warnf("Last call response status: %s", r.StatusCode)
			}
			break
		default:
			time.Sleep(time.Second)
		}
	}
}
