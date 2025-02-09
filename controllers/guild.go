package controllers

import (
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/services"
)

type Guild struct {
	logger   services.Logger
	database data.IDatabase
}

func NewGuildController(logger services.Logger, database data.IDatabase) *Guild {
	return &Guild{logger, database}
}

func (g *Guild) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("GET", "/api/guild", g.getAllGuilds)
}

func (g *Guild) getAllGuilds(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	guilds, err := g.database.GetJoinedGuilds(r.Context(), claims.ID)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, guilds); err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}
