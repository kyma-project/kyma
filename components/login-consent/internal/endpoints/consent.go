package endpoints

import "net/http"

func (cfg *Config) Consent(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("consent"))
}
