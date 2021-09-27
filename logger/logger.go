package logger

import (
	"github.com/sirupsen/logrus"
)

var (
	std *logrus.Logger = logrus.New()
)

// Fields Fields
type Fields map[string]interface{}

func WithField(key string, value interface{}) *logrus.Entry {
	return std.WithField(key, value)
}

func WithFields(fields Fields) *logrus.Entry {
	fie := logrus.Fields(fields)
	return std.WithFields(fie)
}

func Trace(args ...interface{}) {
	std.Trace(args...)
}

// Tracef logs a message at level Trace on the standard logger.
func Tracef(format string, args ...interface{}) {
	std.Tracef(format, args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	std.Error(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	std.Info(args...)
}

func Infoln(args ...interface{}) {
	std.Infoln(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	std.Warn(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}

// Entry Entry
type Entry *logrus.Entry

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *logrus.Entry {
	return std.WithField(logrus.ErrorKey, err)
}
