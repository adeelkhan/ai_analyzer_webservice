package logger

import (
	"log"
	"os"
	"time"
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		warnLogger:  log.New(os.Stderr, "[WARN] ", log.LstdFlags),
	}
}

func GetLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger()
	}
	return defaultLogger
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.warnLogger.Printf(format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.warnLogger.Printf(format, v...)
}

func Info(format string, v ...interface{}) {
	GetLogger().Info(format, v...)
}

func Error(format string, v ...interface{}) {
	GetLogger().Error(format, v...)
}

func Warn(format string, v ...interface{}) {
	GetLogger().Warn(format, v...)
}

func Infof(format string, v ...interface{}) {
	GetLogger().Infof(format, v...)
}

func Errorf(format string, v ...interface{}) {
	GetLogger().Errorf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	GetLogger().Warnf(format, v...)
}

// RequestLogger middleware for HTTP requests
func RequestLogger() func(start time.Time, params ...interface{}) {
	return func(start time.Time, params ...interface{}) {
		duration := time.Since(start)
		GetLogger().Info("Request completed in %v - %s", duration, params)
	}
}
