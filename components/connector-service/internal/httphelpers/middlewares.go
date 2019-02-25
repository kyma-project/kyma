package httphelpers

import "github.com/gorilla/mux"

// WithMiddlewares decorates router with middlewares
func WithMiddlewares(router *mux.Router, middlewares ...mux.MiddlewareFunc) {
	for _, middleware := range middlewares {
		router.Use(middleware)
	}
}
