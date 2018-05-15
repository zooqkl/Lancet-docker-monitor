/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package lancetlogging

import (
	"github.com/op/go-logging"
	"io"
	"os"
)

const (
	pkgLogID      = "lancetlogging"
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{shortfile}] [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x} %{message}%{color:reset}"
	defaultLevel  = logging.ERROR
)

var (
	defaultOutput  = os.Stderr
	loggerInstance *LoggerImpl
)

func GetLogger() Logger {
	if loggerInstance == nil {
		InitLogger("DEBUG")
		loggerInstance.logger.Debug("Init logger with DEBUG level before lancet call InitLogger method and set level to config level.")
	}
	return loggerInstance
}

// SetFormat sets the logging format.
func setFormat(formatSpec string) logging.Formatter {
	if formatSpec == "" {
		formatSpec = defaultFormat
	}
	return logging.MustStringFormatter(formatSpec)
}

func InitLogger(logLevel string) Logger {
	if loggerInstance == nil {
		loggerInstance = &LoggerImpl{
			logger: logging.MustGetLogger(pkgLogID),
		}
		// logger is wrapped in loggerInstance,set ExtraCalldepth=1 to print calldepth where logger.XXX() is called
		loggerInstance.logger.ExtraCalldepth = 1
	} else {
		loggerInstance.logger.Warning("[InitLogger] lancetlogging.InitLogger() has been called more than one times!")
	}

	var level logging.Level
	var err error

	level, err = logging.LogLevel(logLevel)
	if err != nil {
		loggerInstance.Errorf("[GetLogger] Get log level from config error,set level to ERROR: %s", err)
	}

	initBackend(setFormat(defaultFormat), defaultOutput, level)
	return loggerInstance
}

// InitBackend sets up the logging backend based on
// the provided logging formatter and I/O writer.
func initBackend(formatter logging.Formatter, output io.Writer, logLevel logging.Level) {

	backend := logging.NewLogBackend(output, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, formatter)
	// set default log level to high level
	logging.SetBackend(backendFormatter).SetLevel(logging.WARNING, "")
	// and set lancet log level
	logging.SetLevel(logLevel, pkgLogID)
}

type LoggerImpl struct {
	logger *logging.Logger
}

func (l *LoggerImpl) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *LoggerImpl) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LoggerImpl) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *LoggerImpl) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LoggerImpl) Notice(args ...interface{}) {
	l.logger.Notice(args...)
}

func (l *LoggerImpl) Noticef(format string, args ...interface{}) {
	l.logger.Noticef(format, args...)
}

func (l *LoggerImpl) Warn(args ...interface{}) {
	l.logger.Warning(args...)
}

func (l *LoggerImpl) Warnf(format string, args ...interface{}) {
	l.logger.Warningf(format, args...)
}

func (l *LoggerImpl) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *LoggerImpl) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *LoggerImpl) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *LoggerImpl) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
