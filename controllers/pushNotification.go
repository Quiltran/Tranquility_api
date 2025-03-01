package controllers

import (
	"encoding/json"
	"fmt"
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
	app.AddSecureRoute("DELETE", "/api/subscribe", p.unsubscribe)
}

func (p *PushNotificationController) subscribe(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}
	p.logger.INFO(fmt.Sprintf("%s is registering for notifications", claims.Username))

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

func (p *PushNotificationController) unsubscribe(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusUnauthorized, "error")
		return
	}
	p.logger.INFO(fmt.Sprintf("%s is unregistering for notifications", claims.Username))

	if err := p.db.DeleteUserPushInformation(r.Context(), claims.ID); err != nil {
		handleError(w, r, p.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}
}
