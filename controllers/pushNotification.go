package controllers

import (
	"encoding/json"
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/services"

	"github.com/SherClockHolmes/webpush-go"
)

type PushNotificationController struct {
	logger services.Logger
	db     data.IDatabase
}

func NewPushNotificationController(logger services.Logger, db data.IDatabase) *PushNotificationController {
	return &PushNotificationController{
		logger,
		db,
	}
}

func (p *PushNotificationController) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("POST", "/api/subscribe", p.subscribe)
}

func (p *PushNotificationController) subscribe(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}

	var sub webpush.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusBadGateway, "warning")
		return
	}

	if err := p.db.SaveUserPushInformation(r.Context(), &sub, claims.ID); err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

}
