package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/models"
	"tranquility/services"
)

type Auth struct {
	logger   *services.Logger
	database data.IDatabase
}

func NewAuthController(logger *services.Logger, dbCommands data.IDatabase) *Auth {
	return &Auth{logger, dbCommands}
}

func (a *Auth) RegisterRoutes(app *app.App) {
	app.AddRoute("POST", "/api/auth/login", a.login)
	app.AddRoute("POST", "/api/auth/register", a.register)
	app.AddSecureRoute("POST", "/api/auth/refresh", a.refreshToken)
}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	body, err := getJsonBody[models.AuthUser](r)
	if err != nil {
		a.logger.ERROR(fmt.Sprintf("error parsing request body: %v", err))
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	credentials, err := a.database.Login(r.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			a.logger.WARNING(fmt.Errorf("an invalid login attempt occurred for %s: %v", body.Username, err).Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		case errors.Is(err, data.ErrMissingPassword):
			a.logger.WARNING(err.Error())
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		default:
			a.logger.ERROR(err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("content-type", "application/json")
	if err = json.NewEncoder(w).Encode(credentials); err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (a *Auth) register(w http.ResponseWriter, r *http.Request) {
	body, err := getJsonBody[models.AuthUser](r)
	if err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	user, err := a.database.Register(r.Context(), body)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			a.logger.WARNING(fmt.Sprintf("an invalid register request has been made: %v", err))
			http.Error(w, "Invalid Body", http.StatusBadRequest)
			return
		default:
			a.logger.ERROR(fmt.Sprintf("an error occurred while registering user: %v", err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("content-type", "application/json")
	if err = json.NewEncoder(w).Encode(user); err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (a *Auth) refreshToken(w http.ResponseWriter, r *http.Request) {
	var body models.AuthUser
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	requestId, err := getRequestID(r)
	if err != nil {
		a.logger.ERROR(err.Error())
	}

	claims, err := getClaims(r)
	if err != nil {
		a.logger.WARNING(fmt.Sprintf("a request to refresh auth tokens did not have claims in the request: %s", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body.ID = claims.ID
	user, err := a.database.RefreshToken(r.Context(), &body)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			a.logger.WARNING(fmt.Sprintf("a request was made to refresh auth token invalid data. request id: %s", requestId))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		default:
			a.logger.ERROR(fmt.Sprintf("an error occurred while refreshing %d auth token: %v. request id: %s", user.ID, err, requestId))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("content-type", "application/json")
	if err = json.NewEncoder(w).Encode(user); err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
