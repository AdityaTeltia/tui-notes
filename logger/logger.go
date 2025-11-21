package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	level  string
	logger *log.Logger
	file   *os.File
}

var defaultLogger *Logger

func Init(level, logFile string) error {
	var writers []io.Writer
	writers = append(writers, os.Stderr)
	
	var file *os.File
	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		file = f
		writers = append(writers, f)
	}
	
	multiWriter := io.MultiWriter(writers...)
	logger := log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)
	
	defaultLogger = &Logger{
		level:  level,
		logger: logger,
		file:   file,
	}
	
	return nil
}

func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

func shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}
	
	currentLevel := levels[defaultLogger.level]
	msgLevel := levels[level]
	return msgLevel >= currentLevel
}

func Debug(format string, v ...interface{}) {
	if defaultLogger != nil && shouldLog("debug") {
		defaultLogger.logger.Printf("[DEBUG] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if defaultLogger != nil && shouldLog("info") {
		defaultLogger.logger.Printf("[INFO] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if defaultLogger != nil && shouldLog("warn") {
		defaultLogger.logger.Printf("[WARN] "+format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if defaultLogger != nil && shouldLog("error") {
		defaultLogger.logger.Printf("[ERROR] "+format, v...)
	}
}

func Fatal(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.logger.Fatalf("[FATAL] "+format, v...)
	} else {
		log.Fatalf("[FATAL] "+format, v...)
	}
}

func LogRequest(username, action string, err error) {
	if err != nil {
		Error("User: %s, Action: %s, Error: %v", username, action, err)
	} else {
		Info("User: %s, Action: %s", username, action)
	}
}

func LogConnection(username, remoteAddr string) {
	Info("Connection: user=%s, remote=%s", username, remoteAddr)
}

func LogDisconnection(username string, duration time.Duration) {
	Info("Disconnection: user=%s, duration=%v", username, duration)
}

