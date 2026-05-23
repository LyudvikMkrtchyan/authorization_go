package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type WarningLogger interface {
	NoSpace(message string) error
}

type FileWarningLogger struct {
	logDir string
}

func NewFileWarningLogger(logDir string) *FileWarningLogger {
	return &FileWarningLogger{logDir: logDir}
}

func (l *FileWarningLogger) NoSpace(message string) error {
	dir := l.logDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("NoSpace: getwd: %w", err)
		}
		dir = wd
	}

	now := time.Now()
	filename := fmt.Sprintf("warning_%s.txt", now.Format("2006-01-02_15-04-05"))
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("NoSpace: create file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "[%s] WARNING: %s\n", now.Format("2006-01-02 15:04:05"), message)
	if err != nil {
		return fmt.Errorf("NoSpace: write: %w", err)
	}
	return nil
}
