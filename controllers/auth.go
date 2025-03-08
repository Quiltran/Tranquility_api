package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"tranquility/app"
	"tranquility/data"
	"tranquility/models"
	"tranquility/services"
)

type Auth struct {
	logger   services.Logger
	database data.IDatabase
}

func NewAuthController(logger services.Logger, dbCommands data.IDatabase) *Auth {
	return &Auth{logger, dbCommands}
}

func (a *Auth) RegisterRoutes(app *app.App) {
	app.AddRoute("POST", "/api/auth/login", a.login)
	app.AddRoute("POST", "/api/auth/register", a.register)
	app.AddValidatedRoute("POST", "/api/auth/refresh", a.refreshToken)
	app.AddSecureRoute("POST", "/api/webauthn/register/begin", a.beginRegistration)
	app.AddSecureRoute("POST", "/api/webauthn/register/complete", a.completeRegistration)
	app.AddRoute("POST", "/api/webauthn/login/begin", a.beginLogin)
	app.AddRoute("POST", "/api/webauthn/login/complete", a.completeLogin)
}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	body, err := getJsonBody[models.AuthUser](r)
	if err != nil {
		handleError(w, r, a.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}

	user, err := a.database.Login(r.Context(), body, r.RemoteAddr)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			handleError(w, r, a.logger, err, nil, http.StatusUnauthorized, "warning")
			return
		case errors.Is(err, data.ErrMissingPassword):
			handleError(w, r, a.logger, err, nil, http.StatusBadRequest, "warning")
			return
		default:
			handleError(w, r, a.logger, err, nil, http.StatusInternalServerError, "error")
			return
		}
	}

	if err = writeJsonBody(w, user); err != nil {
		handleError(w, r, a.logger, fmt.Errorf("error while logging in: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}

func (a *Auth) register(w http.ResponseWriter, r *http.Request) {
	body, err := getJsonBody[models.AuthUser](r)
	if err != nil {
		handleError(w, r, a.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}

	user, err := a.database.Register(r.Context(), body, r.RemoteAddr)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidCredentials):
			handleError(w, r, a.logger, fmt.Errorf("an invalid register request has been made: %v", err), nil, http.StatusBadRequest, "warning")
			return
		case errors.Is(err, data.ErrInvalidPasswordFormat):
			handleError(w, r, a.logger, fmt.Errorf("a register request has been made with invalid password: %v", err), nil, http.StatusUnauthorized, "warning")
			return
		default:
			handleError(w, r, a.logger, fmt.Errorf("an error occurred while registering user: %v", err), nil, http.StatusInternalServerError, "error")
			return
		}
	}

	if err = writeJsonBody(w, user); err != nil {
		handleError(w, r, a.logger, fmt.Errorf("error while responding to register: %v", err), nil, http.StatusInternalServerError, "error")
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
		case errors.Is(err, data.ErrInvalidCredentials) || errors.Is(err, sql.ErrNoRows):
			a.logger.WARNING(fmt.Sprintf("a request was made to refresh auth token invalid data. request id: %s", requestId))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		default:
			a.logger.ERROR(fmt.Sprintf("an error occurred while refreshing %d auth token: %v. request id: %s", user.ID, err, requestId))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if err = writeJsonBody(w, user); err != nil {
		a.logger.ERROR(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (a *Auth) beginRegistration(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while getting claims to begin webauthn registration: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
	if claims.UserHandle == "" {
		handleError(w, r, a.logger, fmt.Errorf("user_handle was not provided while attempting to register for webauthn"), nil, http.StatusUnauthorized, "error")
		return
	}

	options, err := a.database.RegisterUserWebAuthn(r.Context(), claims)
	if err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while registering user to webauthn: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}

	if err := writeJsonBody(w, options); err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while writing webauthn response to the body: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}

func (a *Auth) completeRegistration(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while getting claims to complete webauthn registration: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}

	if err := a.database.CompleteWebauthnRegister(r.Context(), claims, r); err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while compliting registration for user to webauthn: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}

func (a *Auth) beginLogin(w http.ResponseWriter, r *http.Request) {
	sessionId, options, err := a.database.BeginWebAuthnLogin(r.Context())
	if err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while registering user to webauthn: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "webauthnSession",
		Value:    sessionId,
		Expires:  time.Now().Add(time.Minute),
		Path:     "/",
		SameSite: http.SameSiteNoneMode,
		Domain:   ".quiltran.com",
		Secure:   true,
		HttpOnly: true,
	})

	if err := writeJsonBody(w, options); err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while writing webauthn login response to the body: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}

func (a *Auth) completeLogin(w http.ResponseWriter, r *http.Request) {
	sessionId, err := r.Cookie("webauthnSession")
	if err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred getting session id cookie while completing webauthn login: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
	user, err := a.database.CompleteWebAuthnLogin(r.Context(), sessionId.Value, r)
	if err != nil {
		handleError(w, r, a.logger, fmt.Errorf("an error occurred while compliting registration for user to webauthn: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "webauthnSession",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		},
	)
	if err = writeJsonBody(w, user); err != nil {
		handleError(w, r, a.logger, fmt.Errorf("error while logging in: %v", err), nil, http.StatusInternalServerError, "error")
		return
	}
}
