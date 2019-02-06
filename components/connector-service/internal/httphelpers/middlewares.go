package httphelpers

import "github.com/gorilla/mux"

// WithMiddlewares decorates router with middlewares
func WithMiddlewares(r *mux.Router, middlewares []mux.MiddlewareFunc) {
	for _, middleware := range middlewares {
		r.Use(middleware)
	}
}
