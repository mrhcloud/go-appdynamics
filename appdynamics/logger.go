package appdynamics

import (
	"log"
	"os"
)

type LogLevelType uint

func LogLevel(l LogLevelType) *LogLevelType {
	return &l
}

func (l *LogLevelType) Value() LogLevelType {
	if l != nil {
		return *l
	}
	return LogOff
}

func (l *LogLevelType) Matches(v LogLevelType) bool {
	c := l.Value()
	return c&v == v
}

func (l *LogLevelType) AtLeast(v LogLevelType) bool {
	c := l.Value()
	return c >= v
}

const (
	LogOff LogLevelType = iota * 0x1000
	LogDebug
)

const (
	LogDebugWithSigning LogLevelType = LogDebug | (1 << iota)
	LogDebugWithHTTPBody
	LogDebugWithRequestRetries
	LogDebugWithRequestErrors
	LogDebugWithEventStreamBody
)

type Logger interface {
	Log(...interface{})
}

type LoggerFunc func(...interface{})

func (f LoggerFunc) Log(args ...interface{}) {
	f(args...)
}

func NewDefaultLogger() Logger {
	return &defaultLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

type defaultLogger struct {
	logger *log.Logger
}

func (l defaultLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}