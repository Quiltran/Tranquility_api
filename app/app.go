package app

import (
	"fmt"
	"net/http"
	"tranquility/data"
)

type App struct {
	database data.IDatabase
	mux      *http.ServeMux
}

func CreateApp(database data.IDatabase) App {
	return App{
		database: database,
		mux:      http.NewServeMux(),
	}
}

func (a *App) AddRoute(method string, path string, handler http.HandlerFunc) {
	a.mux.Handle(fmt.Sprintf("%s %s", method, path), handler)
}

func (a App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
