// Package logger is for logger configuration
package logger

import (
	"go.uber.org/zap"
)

var (
	sugarLogger *zap.SugaredLogger
)

// InitLogger initializes both a global zap.Logger and a global zap.SugaredLogger
func InitLogger() {
	logger := zap.Must(zap.NewDevelopment())
	zap.ReplaceGlobals(logger)

	sugarLogger = logger.Sugar()
}

func GetSugaredLogger() *zap.SugaredLogger {
	return sugarLogger
}
