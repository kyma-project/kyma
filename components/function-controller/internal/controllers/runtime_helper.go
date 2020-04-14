package controllers

type RuntimeHelper interface {
	// returns a name of a secret used for pulling images used to serve
	// lambda functions
	Secret() string
	// returns a name of the ServiceAccount to use to run pods serving lambda
	// functions
	ServiceAccount() string
	// returns a name of ConfigMap containing cofiguration of runtime
	RuntimeConfigmap() string
}

var _ RuntimeHelper = &rtHelper{}

type rtHelper struct {
	secret           string
	serviceAccount   string
	runtimeConfigmap string
}

func (h *rtHelper) Secret() string {
	return h.secret
}

func (h *rtHelper) ServiceAccount() string {
	return h.serviceAccount
}

func (h *rtHelper) RuntimeConfigmap() string {
	return h.runtimeConfigmap
}
