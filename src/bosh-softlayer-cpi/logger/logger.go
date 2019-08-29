package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"unsafe"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
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
	GetBoshLogger() boshlog.Logger
	GetSerialTagPrefix() string
	ChangeRetryStrategyLogTag(retryStrategy *boshretry.RetryStrategy) error
}

type logger struct {
	boshlogger   boshlog.Logger
	threadPrefix string
}

func New(level boshlog.LogLevel, serialTagPrefix string, out, err *log.Logger) Logger {
	Logger := boshlog.New(level, out)
	return &logger{
		boshlogger:   Logger,
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

func (l *logger) GetBoshLogger() boshlog.Logger {
	return l.boshlogger
}

func (l *logger) GetSerialTagPrefix() string {
	return l.threadPrefix
}

func (l *logger) Debug(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.boshlogger.Debug(tag, msg, args...)
}

func (l *logger) DebugWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Debug(tag, msg, args...)
}

func (l *logger) Info(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.boshlogger.Info(tag, msg, args...)
}

func (l *logger) Warn(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.boshlogger.Warn(tag, msg, args...)
}

func (l *logger) Error(tag, msg string, args ...interface{}) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.boshlogger.Error(tag, msg, args...)
}

func (l *logger) ErrorWithDetails(tag, msg string, args ...interface{}) {
	msg = msg + "\n********************\n%s\n********************"
	l.Error(tag, msg, args...)
}

func (l *logger) HandlePanic(tag string) {
	tag = fmt.Sprintf("%s:%s", l.threadPrefix, tag)
	l.boshlogger.HandlePanic(tag)
}

func (l *logger) Flush() error { return l.boshlogger.Flush() }

// it's unfriendly to change RetryStrategy().logtag
func (l *logger) ChangeRetryStrategyLogTag(retryStrategy *boshretry.RetryStrategy) error {
	defer func() {
		if err := recover(); err != nil {
			l.Warn("cpiLogger", fmt.Sprintf("Recovered from panic when ChangeRetryStrategyLogTag: %s", err))
		}
	}()

	//retryStrategy only refer interface RetryStrategy, so add '*' to get private timeoutRetryStrategy
	pointerVal := reflect.ValueOf(*retryStrategy)
	val := reflect.Indirect(pointerVal)

	logtag := val.FieldByName("logTag")
	// Overwrite unexported bosh-utils/retrystrategy/timeoutRetryStrategy.logger to distinguish logs
	// #nosec G103
	ptrToLogTag := unsafe.Pointer(logtag.UnsafeAddr())
	realPtrToLogTag := (*string)(ptrToLogTag)
	serialTagPrefix := fmt.Sprintf("%s:%s", l.GetSerialTagPrefix(), logtag)
	*realPtrToLogTag = serialTagPrefix

	return nil
}
