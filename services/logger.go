package services

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	INFO    = "\033[36m"
	WARNING = "\033[32m"
	ERROR   = "\033[31m"
	TRACE   = "\033[94m"
	RESET   = "\033[0m"
)

type Logger interface {
	INFO(message string)
	WARNING(message string)
	ERROR(message string)
	TRACE(message string)
}

type ConsoleLogger struct {
	console *log.Logger
	file    *log.Logger
}

func CreateLogger(name string) (*ConsoleLogger, error) {
	file, err := os.OpenFile(name+".log", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while seeking to the end of the log file: %v", err)
	}

	return &ConsoleLogger{
		console: log.New(os.Stdout, "Tranquility ", log.LstdFlags),
		file:    log.New(file, "Tranquility ", log.LstdFlags),
	}, nil
}

func (l *ConsoleLogger) INFO(message string) {
	l.console.Printf("%s[INFO]%s %s\n", INFO, RESET, message)
	l.file.Printf("[INFO] %s\n", message)
}

func (l *ConsoleLogger) WARNING(message string) {
	l.console.Printf("%s[WARNING]%s %s\n", WARNING, RESET, message)
	l.file.Printf("[WARNING] %s\n", message)
}

func (l *ConsoleLogger) ERROR(message string) {
	l.console.Printf("%s[ERROR]%s %s\n", ERROR, RESET, message)
	l.file.Printf("[ERROR] %s\n", message)
}

func (l *ConsoleLogger) TRACE(message string) {
	l.console.Printf("%s[TRACE]%s %s\n", TRACE, RESET, message)
	l.file.Printf("[TRACE] %s\n", message)
}
