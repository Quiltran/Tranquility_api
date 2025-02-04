package controllers

import (
	"fmt"
	"net/http"
	"tranquility/app"
	"tranquility/data"
	"tranquility/services"
)

type Attachment struct {
	logger   *services.Logger
	database data.IDatabase
}

func NewAttachmentController(logger *services.Logger, database data.IDatabase) *Attachment {
	return &Attachment{logger, database}
}

func (a *Attachment) RegisterRoutes(app *app.App) {
	app.AddSecureRoute("POST", "/api/attachment", a.uploadAttachment)
}

func (a *Attachment) uploadAttachment(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1024)
	if err != nil {
		a.logger.ERROR(fmt.Sprintf("an error occurred while collecting form body while uploading attachment: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		a.logger.ERROR(fmt.Sprintf("an upload occurred but errored out when getting file: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileType := handler.Header.Get("Content-Type")
	if fileType == "" {
		a.logger.WARNING(fmt.Sprintf("an upload occurred without a file type: %v", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fileName := handler.Filename
	if fileName == "" {
		a.logger.WARNING(fmt.Sprintf("an upload occurred without a valid file name: %v", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	outputName, outputPath, err := services.StoreFile(&file, fileName)
	if err != nil {
		a.logger.ERROR(fmt.Sprintf("an error occurred while storing file to disk: %v", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Println(outputName, outputPath)
}
