package controllers

import (
	"errors"
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/models"
	"tranquility/services"
)

type Attachment struct {
	logger      services.Logger
	fileHandler *services.FileHandler
	database    data.IDatabase
}

func NewAttachmentController(logger services.Logger, fileHandler *services.FileHandler, database data.IDatabase) *Attachment {
	return &Attachment{logger, fileHandler, database}
}

func (a *Attachment) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("POST", "/api/attachment", a.uploadAttachment)
	app.AddSecureRoute("DELETE", "/api/attachment", a.uploadAttachment)
}

func (a *Attachment) uploadAttachment(w http.ResponseWriter, r *http.Request) {
	claims, err := getClaims(r)
	if err != nil {
		handleError(w, a.logger, err, nil, http.StatusBadRequest, "error")
		return
	}

	err = r.ParseMultipartForm(10 * 1024)
	if err != nil {
		handleError(w, a.logger, err, claims, http.StatusInternalServerError, "error")
		return
	}

	attachment, file, err := models.NewAttachmentFromRequest(r, claims.ID, "file")
	if err != nil {
		code := http.StatusInternalServerError
		level := "ERROR"

		if errors.Is(err, models.ErrAttachmentNoContentType) ||
			errors.Is(err, models.ErrAttachmentNoFileName) {
			code = http.StatusBadRequest
			level = "WARNING"
		}
		handleError(w, a.logger, err, claims, code, level)
		return
	}
	defer file.Close()

	output, err := a.database.CreateAttachment(r.Context(), &file, attachment)
	if err != nil {
		handleError(w, a.logger, err, claims, http.StatusBadRequest, "error")
		return
	}
	if err = writeJsonBody(w, *output); err != nil {
		handleError(w, a.logger, err, claims, http.StatusInternalServerError, "error")
		return
	}
}
