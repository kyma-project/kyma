package app

import "net/http"

func (a App) HandleConsoleRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.busolaURL, http.StatusFound)
}

func (a App) HandleIndexRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}
