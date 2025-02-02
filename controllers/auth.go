package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tranquility/app"
	"tranquility/config"
	"tranquility/models"
	"tranquility/services"
)

type Auth struct {
	logger     *services.Logger
	dbCommands *services.DatabaseCommands
	config     *config.Config
}

func NewAuthController(logger *services.Logger, dbCommands *services.DatabaseCommands, config *config.Config) *Auth {
	return &Auth{logger, dbCommands, config}
}

func (a *Auth) RegisterRoutes(app *app.App) {
	app.AddRoute("POST", "/api/auth/login", a.Login)
	app.AddRoute("POST", "/api/auth/register", a.Register)
}

func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var body models.AuthUser
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	if body.Password == "" {
		a.logger.WARNING("a login request happened without a proper body")
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	credentials, err := a.dbCommands.Login(r.Context(), &body)
	if err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	passwordsMatch, err := services.VerifyPassword(body.Password, credentials.Password)
	if err != nil {
		a.logger.ERROR(fmt.Sprintf("unable to verify password hash: %s", err))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if !passwordsMatch {
		a.logger.WARNING("an invalid login attempt took place: bad password")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	authToken, err := services.GenerateToken(credentials)
	if err != nil {
		a.logger.ERROR(fmt.Errorf("an error occurred while generating token: %v", err).Error())
	}
	credentials.Token = authToken
	credentials.ClearAuth()

	w.Header().Add("content-type", "application/json")
	if err = json.NewEncoder(w).Encode(credentials); err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (a *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var body models.AuthUser
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	if body.Password == "" || body.ConfirmPassword == "" {
		a.logger.WARNING("a register request happened with invalid passwords")
		http.Error(w, "Invalid Body", http.StatusBadRequest)
		return
	}

	body.Password, err = services.HashPassword(body.Password)
	if err != nil {
		a.logger.ERROR(fmt.Errorf("an error occurred hashing password while registering user: %v", err).Error())
	}

	user, err := a.dbCommands.Register(r.Context(), &body)
	if err != nil {
		a.logger.ERROR(fmt.Errorf("an error occurred while registering user: %v", err).Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Add("content-type", "application/json")
	if err = json.NewEncoder(w).Encode(user); err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
