package server

import (
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

const (
	DebugLevel = 1
	InfoLevel = 2
	WarningLevel = 3
	ErrorLevel = 0
)

var once sync.Once
var AppLogger *logrus.Logger


func InitLogger(logLevel int, reportCaller bool) {

	once.Do(func() {

		AppLogger = &logrus.Logger {
			Out:          os.Stdout,
			Hooks:        make(logrus.LevelHooks),
			Formatter:    new(logrus.TextFormatter),
			ReportCaller: reportCaller,
			ExitFunc:     os.Exit,
		}
		if logLevel == DebugLevel {
			AppLogger.Level = logrus.DebugLevel
		} else if logLevel == InfoLevel {
			AppLogger.Level = logrus.InfoLevel
		} else if logLevel == WarningLevel {
			AppLogger.Level = logrus.WarnLevel
		} else {
			AppLogger.Level = logrus.ErrorLevel
		}
	})
}