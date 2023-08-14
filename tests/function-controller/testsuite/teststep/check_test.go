package teststep

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/sirupsen/logrus"
	"net/url"
	"testing"
)

func TestName(t *testing.T) {
	log := logrus.New().WithField("test", "cloud event")
	fnURL, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}

	check := NewCloudEventCheck(cloudevents.EncodingBinary, log, "test", fnURL)

	err = check.Run()
	if err != nil {
		panic(err)
	}
}
