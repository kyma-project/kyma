package app

import (
	"log"
	"net/http"
	"strings"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/renderer"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

func (a App) HandleStaticAssets(w http.ResponseWriter, r *http.Request) {
	// disable assets directory listing
	if strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	rctx := chi.RouteContext(r.Context())
	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	fs := http.StripPrefix(pathPrefix, http.FileServer(a.fsAssets))
	fs.ServeHTTP(w, r)
}

func (a App) HandleStaticIndex(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UAAEnabled bool
	}{
		UAAEnabled: a.UAAEnabled,
	}

	err := a.htmlRenderer.RenderTemplate(w, renderer.TemplateNameIndex, data)
	if err != nil {
		log.Println(errors.Wrapf(err, "while rendering HTML template [%s]", renderer.TemplateNameIndex))
		http.Error(w, "Unable to render webpage", http.StatusInternalServerError)
		return
	}
}

func (a App) HandleStaticSuccess(w http.ResponseWriter, r *http.Request) {
	err := a.htmlRenderer.RenderTemplate(w, renderer.TemplateNameSuccess, nil)
	if err != nil {
		log.Println(errors.Wrapf(err, "while rendering HTML template [%s]", renderer.TemplateNameSuccess))
		http.Error(w, "Unable to render webpage", http.StatusInternalServerError)
		return
	}
}
