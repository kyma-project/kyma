package graphql

import (
	"encoding/json"

	"github.com/machinebox/graphql"
)

type Request struct {
	query string
	vars  map[string]interface{}
	req   *graphql.Request
}

func NewRequest(q string) *Request {
	return &Request{
		query: q,
		req:   graphql.NewRequest(q),
		vars:  make(map[string]interface{}),
	}
}

func (r *Request) SetVar(key string, value interface{}) {
	r.vars[key] = value
	r.req.Var(key, value)
}

func (r *Request) AddHeader(key, value string) {
	r.req.Header.Add(key, value)
}

func (r *Request) Json() ([]byte, error) {
	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     r.query,
		Variables: r.vars,
	}

	return json.Marshal(requestBodyObj)
}
