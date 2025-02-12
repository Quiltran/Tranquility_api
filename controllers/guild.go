package controllers

import (
	"database/sql"
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
	app.AddSecureRoute("GET", "/api/guild/{guildId}/channel", g.getGuildChannels)
	app.AddSecureRoute("GET", "/api/guild/{guildId}/channel/{channelId}", g.getGuildChannel)
}

func (g *Guild) getAllGuilds(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	guilds, err := g.database.GetJoinedGuilds(r.Context(), claims.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
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
		if err == sql.ErrNoRows {
			handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
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
		if err == sql.ErrNoRows {
			handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, guild); err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}

func (g *Guild) getGuildChannels(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	pathGuildID := r.PathValue("guildId")
	if pathGuildID == "" {
		handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}
	guildId, err := strconv.Atoi(pathGuildID)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	channels, err := g.database.GetGuildChannels(r.Context(), int32(guildId), claims.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, channels); err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}

func (g *Guild) getGuildChannel(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	pathGuildID := r.PathValue("guildId")
	pathChannelID := r.PathValue("channelId")
	if pathGuildID == "" || pathChannelID == "" {
		handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}

	guildId, err := strconv.Atoi(pathGuildID)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
	channelId, err := strconv.Atoi(pathChannelID)
	if err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	channel, err := g.database.GetGuildChannel(r.Context(), int32(guildId), int32(channelId), claims.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(w, g.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, channel); err != nil {
		handleError(w, g.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}
