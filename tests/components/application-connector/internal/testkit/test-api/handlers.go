package test_api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

// resCode should only be used in paths with `code`
// parameter, that is a valid int
func resCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	codeStr := vars["code"]          // must exist, because path has a pattern
	code, _ := strconv.Atoi(codeStr) // can't error, because path has a pattern
	w.WriteHeader(code)
	w.Write([]byte(codeStr))
}

func timeout(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Done()
	if c == nil {
		log.Println("Context has no timeout, sleeping for 2 minutes")
		time.Sleep(2 * time.Minute)
		return
	}
	log.Println("Context timeout, waiting until done")

	_ = <-c

	alwaysOk(w, r)
}
