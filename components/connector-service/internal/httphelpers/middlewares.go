package httphelpers

import "github.com/gorilla/mux"

// WithMiddlewares decorates router with middlewares
func WithMiddlewares(middlewares []mux.MiddlewareFunc, routers ...*mux.Router) {
	for _, r := range routers {
		for _, middleware := range middlewares {
			r.Use(middleware)
		}
	}
}
