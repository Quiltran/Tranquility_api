package controllers

import (
	"net/http"
	"strconv"
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
	app.AddSecureRoute("GET", "/api/guild/owned", g.getOwnedGuilds)
	app.AddSecureRoute("GET", "/api/guild/{guildId}", g.getGuild)
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

func (g *Guild) getOwnedGuilds(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	guilds, err := g.database.GetOwnedGuilds(r.Context(), claims.ID)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, guilds); err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}

func (g *Guild) getGuild(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	pathGuildId := r.PathValue("guildId")
	if pathGuildId == "" {
		handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}
	guildId, err := strconv.Atoi(pathGuildId)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	guild, err := g.database.GetGuildByID(r.Context(), int32(guildId), claims.ID)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, guild); err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}
