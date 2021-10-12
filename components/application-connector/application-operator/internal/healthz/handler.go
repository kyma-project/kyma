package healthz

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func newHTTPHandler(log *logrus.Logger) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	}
}
