package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Logger interface {
	Debug(tag, msg string, args ...interface{})
	DebugWithDetails(tag, msg string, args ...interface{})
	Info(tag, msg string, args ...interface{})
	Warn(tag, msg string, args ...interface{})
	Error(tag, msg string, args ...interface{})
	ErrorWithDetails(tag, msg string, args ...interface{})
	HandlePanic(tag string)
	Flush() error
	GetBasicLogger() boshlog.Logger
	GetSerialTagPrefix() string
}

type logger struct {
	basicLogger  boshlog.Logger
	threadPrefix string
}

func New(level boshlog.LogLevel, serialTagPrefix string, out, err *log.Logger) Logger {
	Logger := boshlog.New(level, out, err)
	return &logger{
		basicLogger:  Logger,
		threadPrefix: serialTagPrefix,
	}
}

func NewLogger(level boshlog.LogLevel, serialTagPrefix string) Logger {
	return NewWriterLogger(level, serialTagPrefix, os.Stdout, os.Stderr)
}

func NewWriterLogger(level boshlog.LogLevel, serialTagPrefix string, out, err io.Writer) Logger {
	return New(
		level,
		serialTagPrefix,
		log.New(out, "", log.LstdFlags),
		log.New(err, "", log.LstdFlags),
	)
}

func (l *logger) GetBasicLogger() boshlog.Logger {
	return l.basicLogger
}

func (l *logger) GetSerialTagPrefix() string {
	return l.threadPrefix
}

func (l *logger) Debug(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.basicLogger.Debug(tag, msg, args...)
}

func (l *logger) DebugWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Debug(tag, msg, args...)
}

func (l *logger) Info(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.basicLogger.Info(tag, msg, args...)
}

func (l *logger) Warn(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.basicLogger.Warn(tag, msg, args...)
}

func (l *logger) Error(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s;%s", l.threadPrefix, tag)
	l.basicLogger.Error(tag, msg, args...)
}

func (l *logger) ErrorWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Error(tag, msg, args...)
}

func (l *logger) HandlePanic(tag string) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.basicLogger.HandlePanic(tag)
}

func (l *logger) Flush() error { return l.basicLogger.Flush() }
