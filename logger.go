package brisk

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

var AppLogger Logger
var once sync.Once

type Logger struct {
	logger *logrus.Logger
}

func InitLogger(logLevel int, reportCaller bool) {

	once.Do(func() {

		AppLogger = Logger{logger: &logrus.Logger {
			Out:          os.Stdout,
			Hooks:        make(logrus.LevelHooks),
			Formatter:    new(logrus.TextFormatter),
			ReportCaller: reportCaller,
			ExitFunc:     os.Exit,
		}}
		if logLevel == DebugLevel {
			AppLogger.logger.Level = logrus.DebugLevel
		} else if logLevel == InfoLevel {
			AppLogger.logger.Level = logrus.InfoLevel
		} else if logLevel == WarningLevel {
			AppLogger.logger.Level = logrus.WarnLevel
		} else {
			AppLogger.logger.Level = logrus.ErrorLevel
		}
	})
}