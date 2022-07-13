package test_api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func NewEchoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		extractQueryRequest(w, r)
		extractHeaders(w, r)
		extractBody(w, r)
	}
}

func extractQueryRequest(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s --> %s\n", req.Method, req.URL)
}

func extractHeaders(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "%s: %s\n", k, v)
	}
}

func extractBody(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		panic(err)
	}

	bodyStr := fmt.Sprintf("body:\n %s\n", body)

	fmt.Println(bodyStr)
	io.WriteString(w, bodyStr)
}
