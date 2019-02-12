package httphelpers

import "net/http"

type WriterWithStatus struct {
	http.ResponseWriter
	status int
}

func (w *WriterWithStatus) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *WriterWithStatus) IsSuccessful() bool {
	return w.status == http.StatusOK || w.status == http.StatusCreated
}
