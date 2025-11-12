package utils

import (
	"log"
	"os"
)

const loggerFlags = log.Ldate | log.Ltime | log.Lshortfile

type CustomLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

var instance *CustomLogger

// Gets the singleton instance of the logger
func GetCustomLogger() *CustomLogger {

	if instance == nil {
		instance = &CustomLogger{
			infoLogger:  log.New(os.Stdout, "[INFO]\t", loggerFlags),
			errorLogger: log.New(os.Stderr, "[ERROR]\t", loggerFlags),
		}
	}

	return instance
}

func (logger *CustomLogger) Info(log any) {
	instance.infoLogger.Println(log)
}

func (logger *CustomLogger) Infof(format string, values ...any) {
	instance.infoLogger.Printf(format, values...)
}

func (logger *CustomLogger) Error(log string) {
	instance.errorLogger.Println(log)
}

func (logger *CustomLogger) Errorf(format string, values ...any) {
	instance.errorLogger.Printf(format, values...)
}
