package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Logger - logger interface that tgf expects from the user.
type Logger interface {
	Info(string, ...any)
	Error(string, ...any)
	Warn(string, ...any)
	Debug(string, ...any)
}

// DefaultLogger - This is a basic logger that tgf comes with in case you do not want to wrap your own logger.
type DefaultLogger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	debugLogger   *log.Logger
	errorLogger   *log.Logger
	debug         bool
}

const (
	reset  = "\x1b[0m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	blue   = "\x1b[34m"
)

// NewDefaultLogger - returns an instance of the default built in logger.
func NewDefaultLogger(debug bool) *DefaultLogger {
	prefix := "[HRP]: "

	infoLogger := log.New(os.Stdout, green+prefix, 0)
	warningLogger := log.New(os.Stdout, yellow+prefix, 0)
	debugLogger := log.New(os.Stdout, blue+prefix, 0)
	errorLogger := log.New(os.Stderr, red+prefix, 0)

	return &DefaultLogger{
		infoLogger:    infoLogger,
		warningLogger: warningLogger,
		debugLogger:   debugLogger,
		errorLogger:   errorLogger,
		debug:         debug,
	}
}

func (l *DefaultLogger) getFileName() string {
	_, file, line, _ := runtime.Caller(2)
	fileName := filepath.Base(file)
	return fmt.Sprintf("[%s:%d]: ", fileName, line)
}

func (l *DefaultLogger) Info(format string, v ...any) {
	l.infoLogger.Printf(reset+format, v...)
}

func (l *DefaultLogger) Error(format string, v ...any) {
	source := l.getFileName()
	l.errorLogger.Printf("[ERROR] "+source+format+reset, v...)
}

func (l *DefaultLogger) Warn(format string, v ...any) {
	l.warningLogger.Printf("[WARN]: "+format+reset, v...)
}

func (l *DefaultLogger) Debug(format string, v ...any) {
	if !l.debug {
		return
	}
	soruce := l.getFileName()
	l.debugLogger.Printf("[DEBUG] "+soruce+format+reset, v...)
}
