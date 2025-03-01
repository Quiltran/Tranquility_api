package controllers

import (
	"fmt"
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/services"
)

type profileController struct {
	logger services.Logger
	db     data.IDatabase
}

func NewProfileController(logger services.Logger, db data.IDatabase) *profileController {
	return &profileController{
		logger,
		db,
	}
}

func (p *profileController) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("GET", "/api/profile", p.GetProfile)
}

func (p *profileController) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	profile, err := p.db.GetUserProfile(r.Context(), claims.ID)
	if err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	if err = writeJsonBody(w, profile); err != nil {
		handleError(w, r, p.logger, fmt.Errorf("an error occurred while serializing profile information: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}
