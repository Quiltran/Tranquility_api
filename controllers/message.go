package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"tranquility/app"
	"tranquility/data"
	"tranquility/services"
)

type Message struct {
	logger   services.Logger
	database data.IDatabase
}

func NewMessageController(logger services.Logger, database data.IDatabase) *Message {
	return &Message{logger, database}
}

func (m *Message) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("GET", "/api/guild/{guildId}/channel/{channelId}/message/page/{pageNumber}", m.getChannelMessages)
}

func (m *Message) getChannelMessages(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, m.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	guildId, guildErr := strconv.ParseInt(r.PathValue("guildId"), 10, 32)
	channelId, channelErr := strconv.ParseInt(r.PathValue("channelId"), 10, 32)
	pageNumber, pageNumErr := strconv.ParseInt(r.PathValue("pageNumber"), 10, 32)
	if guildErr != nil || channelErr != nil || pageNumErr != nil {
		handleError(w, m.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}

	messages, err := m.database.GetChannelMessages(
		r.Context(),
		claims.ID,
		int32(guildId),
		int32(channelId),
		int32(pageNumber),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(w, m.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
		handleError(w, m.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, messages); err != nil {
		handleError(w, m.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}
