package app

import (
	"fmt"
	"net/http"
	"tranquility/data"
	"tranquility/middleware"
	"tranquility/services"
)

type App struct {
	database data.IDatabase
	mux      *http.ServeMux
	logger   *services.Logger
}

func CreateApp(database data.IDatabase, logger *services.Logger) App {
	return App{
		database: database,
		mux:      http.NewServeMux(),
		logger:   logger,
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
