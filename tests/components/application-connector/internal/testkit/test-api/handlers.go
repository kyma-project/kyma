package test_api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func alwaysOk(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type EchoResponse struct {
	Body    []byte              `json:"body"`
	Headers map[string][]string `json:"headers"`
	Method  string              `json:"method"`
	Query   string              `json:"query"`
}

func echo(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Couldn't read request body:", r.URL)
		body = []byte("<failed to read>")
	}

	res := EchoResponse{
		Method:  r.Method,
		Body:    body,
		Headers: r.Header,
		Query:   r.URL.RawQuery,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		log.Println("Couldn't encode the response body to JSON:", r.URL)
	}
}
