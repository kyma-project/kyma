package graphql

import "github.com/machinebox/graphql"

type Request struct {
	req *graphql.Request
}

func NewRequest(q string) *Request {
	return &Request{
		req: graphql.NewRequest(q),
	}
}

func (r *Request) SetVar(key string, value interface{}) {
	r.req.Var(key, value)
}

func (r *Request) AddHeader(key, value string) {
	r.req.Header.Add(key, value)
}
