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
		return "", "", fmt.Errorf("an error occurred while opening output file: %v", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, *file)
	if err != nil {
		return "", "", fmt.Errorf("an error occurred while copying temp file into file: %v", err)
	}

	return fileName, filePath, nil
}

func (f *FileHandler) GetFileUrl(fileName string) (string, error) {
	filePath := path.Join(f.uploadPath, fileName)
	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("the file provided does not exist: %v", err)
		}
		return "", err
	}

	return "/api/attachment/" + fileName, nil
}

func (f *FileHandler) DeleteFile(fileName string) error {
	filePath := path.Join(f.uploadPath, fileName)
	return os.Remove(filePath)
}
