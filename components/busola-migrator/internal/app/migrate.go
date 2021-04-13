package app

import "net/http"

func (a App) HandleXSUAAMigrate(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("not implemented"))
}
