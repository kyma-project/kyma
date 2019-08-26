package service

import "net/http"

func NewTestService(config Config) *service {
	return &service{
		host: config.Host,
		port: config.Port,
	}
}

func (s *service) SetupHandlers() *http.ServeMux {
	return s.setupHandlers()
}
