package assertion

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/sirupsen/logrus"
	"net/url"
	"testing"
)

func TestCloudEventCheckLocally(t *testing.T) {
	t.Run("cloud event check", func(t *testing.T) {
		testCases := map[string]struct {
			cloudevents.Encoding
		}{
			"Structured": {
				cloudevents.EncodingStructured,
			},
			"Binary": {
				cloudevents.EncodingBinary,
			},
		}
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				//GIVEN
				log := logrus.New().WithField("test", "cloud event")
				fnURL, err := url.Parse("http://localhost:8080")
				if err != nil {
					panic(err)
				}

				//WHEN
				check := CloudEventReceiveCheck(log, "test", tc.Encoding, fnURL)

				//THEN
				err = check.Run()
				if err != nil {
					panic(err)
				}
			})
		}

	})

	t.Run("cloud event send check", func(t *testing.T) {
		//GIVEN
		log := logrus.New().WithField("test", "cloud event")
		fnURL, err := url.Parse("http://localhost:8080")
		if err != nil {
			panic(err)
		}

		//WHEN
		check := CloudEventSendCheck(log, "test", fnURL)

		//THEN
		err = check.Run()
		if err != nil {
			panic(err)
		}
	})
}
