package app

import (
	"fmt"
	"net/http"
	"tranquility/middleware"
	"tranquility/services"
)

type App struct {
	mux    *http.ServeMux
	logger services.Logger
}

func CreateApp(logger services.Logger) App {
	return App{
		mux:    http.NewServeMux(),
		logger: logger,
	}
}

func (a *App) AddRoute(method string, path string, handler http.HandlerFunc) {
	a.mux.Handle(fmt.Sprintf("%s %s", method, path), handler)
}
func (a *App) AddSecureRoute(method string, path string, handler http.HandlerFunc) {
	wrappedHandler := middleware.ValidateJWT(handler, a.logger)
	a.mux.Handle(fmt.Sprintf("%s %s", method, path), wrappedHandler)
}

func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
