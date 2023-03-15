package api

import (
	"k8s.io/client-go/dynamic"
)

type Client struct {
	client dynamic.Interface
}

func NewClient(client dynamic.Interface) *Client {
	return &Client{client: client}
}
