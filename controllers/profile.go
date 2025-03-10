package controllers

import (
	"fmt"
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/models"
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
	app.AddSecureRoute("PATCH", "/api/profile", p.UpdateProfile)
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

func (p *profileController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	body, err := getJsonBody[models.Profile](r)
	if err != nil {
		err = fmt.Errorf("an error occurred while collecting body in order to update profile: %v", err)
		handleError(w, r, p.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}

	profile, err := p.db.UpdateUserProfile(r.Context(), body, claims.ID)
	if err != nil {
		err = fmt.Errorf("an error occurred while updating profile: %v", err)
		handleError(w, r, p.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}

	if err = writeJsonBody(w, profile); err != nil {
		handleError(w, r, p.logger, fmt.Errorf("an error occurred while serializing profile information after update: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}
