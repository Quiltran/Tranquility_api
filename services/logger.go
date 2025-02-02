package services

import (
	"log"
	"os"
)

const (
	INFO    = "\033[36m"
	WARNING = "\033[32m"
	ERROR   = "\033[31m"
	RESET   = "\033[0m"
)

type Logger struct {
	console *log.Logger
	file    *log.Logger
}

func CreateLogger(name string) (*Logger, error) {
	file, err := os.OpenFile(name+".log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return nil, err
	}

	return &Logger{
		console: log.New(os.Stdout, "Tranquility ", log.LstdFlags),
		file:    log.New(file, "Tranquility ", log.LstdFlags),
	}, nil
}

func (l *Logger) INFO(message string) {
	l.console.Printf("%s[INFO]%s %s\n", INFO, RESET, message)
	l.file.Printf("[INFO] %s\n", message)
}

func (l *Logger) WARNING(message string) {
	l.console.Printf("%s[WARNING]%s %s\n", WARNING, RESET, message)
	l.file.Printf("[WARNING] %s\n", message)
}

func (l *Logger) ERROR(message string) {
	l.console.Printf("%s[ERROR]%s %s\n", ERROR, RESET, message)
	l.file.Printf("[ERROR] %s\n", message)
}
