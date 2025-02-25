package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"tranquility/app"
	"tranquility/data"
	"tranquility/services"
)

type Member struct {
	logger   services.Logger
	database data.IDatabase
}

func NewMemberController(logger services.Logger, database data.IDatabase) *Member {
	return &Member{logger, database}
}

func (m *Member) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("GET", "/api/member", m.getMembers)
}

func (m *Member) getMembers(w http.ResponseWriter, r *http.Request) {
	_, err := getClaims(r)
	if err != nil {
		handleError(w, r, m.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	guildId, err := strconv.Atoi(r.URL.Query().Get("exclude"))
	if err != nil {
		guildId = -1
	}

	members, err := m.database.GetMembers(r.Context(), int32(guildId))
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(w, r, m.logger, err, nil, http.StatusBadRequest, "warning")
			return
		}
		handleError(w, r, m.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	if err = writeJsonBody(w, members); err != nil {
		handleError(w, r, m.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}
