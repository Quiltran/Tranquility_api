package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"time"
)

var (
	fileUploadPath string = "./uploads"
)

func init() {
	path := os.Getenv("UPLOAD_PATH")
	if path != "" {
		fileUploadPath = path
	}
	if !checkDestination(fileUploadPath) {
		panic("UPLOAD_PATH provided is invalid")
	}
}

func checkDestination(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func StoreFile(file *multipart.File, fileName string) (string, string, error) {
	timestamp := time.Now().Unix()

	fileName = fmt.Sprintf("%d-%s", timestamp, fileName)

	filePath := path.Join(fileUploadPath, fileName)

	outputFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", "", err
	}

	_, err = io.Copy(outputFile, *file)
	if err != nil {
		return "", "", err
	}

	return fileName, filePath, err
}
