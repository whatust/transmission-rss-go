package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/whatust/transmission-rss/config"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// Hook  ...
type Hook struct {
	formatter logrus.Formatter
	logLevel  logrus.Level
	logger    *lumberjack.Logger
}

// NewHook creates a new instance of Hook
func NewHook(opt config.Log) *Hook {

	var level logrus.Level
	var formatter logrus.Formatter

	switch opt.Level {
	case "Debug":
		level = logrus.DebugLevel
	case "Info":
		level = logrus.InfoLevel
	case "Error":
		level = logrus.ErrorLevel
	case "Warning":
		level = logrus.WarnLevel
	default:
		level = logrus.InfoLevel
	}

	switch opt.Formatter {
	case "JSON":
		formatter = &logrus.JSONFormatter{}
	default:
		formatter = &logrus.TextFormatter{DisableColors: true}
	}

	hook := Hook{
		logger: &lumberjack.Logger{
			Filename:   opt.LogPath,
			MaxSize:    opt.MaxSize,
			MaxBackups: opt.MaxBackups,
			MaxAge:     opt.MaxAge,
			Compress:   opt.Compress,
			LocalTime:  opt.LocalTime,
		},
		logLevel:  level,
		formatter: formatter,
	}

	return &hook
}

// Fire function that executes when hook is activated
func (hook *Hook) Fire(entry *logrus.Entry) error {

	msg, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = hook.logger.Write([]byte(msg))

	return err
}

// Levels return levels to activate hook
func (hook *Hook) Levels() []logrus.Level {

	return logrus.AllLevels[:hook.logLevel+1]
}

var log *logrus.Logger

func init() {
	log = logrus.New()
}

// ConfigLogger initialize log parameters
func ConfigLogger(config config.Log) (*logrus.Logger, error) {

	log.SetOutput(os.Stdout)
	//log.SetFormatter(&logrus.JSONFormatter{})

	switch config.Level {
	case "Error":
		log.SetLevel(logrus.ErrorLevel)
	case "Warning":
		log.SetLevel(logrus.WarnLevel)
	case "Info":
		log.SetLevel(logrus.InfoLevel)
	case "Debug":
		log.SetLevel(logrus.DebugLevel)
	}

	hook := NewHook(config)
	log.AddHook(hook)

	return log, nil
}

const (
	// PanicLevel expose logrus log level cosntants
	PanicLevel = logrus.PanicLevel
	// FatalLevel expose logrus log level cosntants
	FatalLevel = logrus.FatalLevel
	// ErrorLevel expose logrus log level cosntants
	ErrorLevel = logrus.ErrorLevel
	// WarnLevel expose logrus log level cosntants
	WarnLevel = logrus.WarnLevel
	// InfoLevel expose logrus log level cosntants
	InfoLevel = logrus.InfoLevel
	// DebugLevel expose logrus log level cosntants
	DebugLevel = logrus.DebugLevel
	// TraceLevel expose logrus log level cosntants
	TraceLevel = logrus.TraceLevel
)

// IsLevelGreaterEqual return true is log level is greater than threshold
func IsLevelGreaterEqual(t logrus.Level) bool {
	return t >= logrus.GetLevel()
}

// Info sends a formated message at level Info to logger
func Info(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Warn sends a formated message at level Warning to logger
func Warn(format string, v ...interface{}) {
	log.Warnf(format, v...)
}

// Error sends a formated message at level Error to logger
func Error(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Debug sends a formated message at level Debug to logger
func Debug(format string, v ...interface{}) {
	log.Debugf(format, v...)
}
