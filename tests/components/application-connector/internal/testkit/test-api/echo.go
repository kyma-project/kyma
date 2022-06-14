package test_api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func NewEchoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: handle Query parameters
		extractHeaders(w, r)
		extractBody(w, r)
	}
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
