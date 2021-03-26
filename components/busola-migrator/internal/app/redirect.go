package app

import "net/http"

func (a App) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.busolaURL, 301)
}
