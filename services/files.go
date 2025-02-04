package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"time"
)

type FileHandler struct {
	uploadPath string
}

func NewFileHandler(path string) *FileHandler {
	return &FileHandler{path}
}

func (f *FileHandler) StoreFile(file *multipart.File, fileName string) (string, string, error) {
	timestamp := time.Now().Unix()

	fileName = fmt.Sprintf("%d-%s", timestamp, fileName)

	filePath := path.Join(f.uploadPath, fileName)

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
