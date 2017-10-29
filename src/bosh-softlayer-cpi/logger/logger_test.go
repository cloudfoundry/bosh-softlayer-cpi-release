package logger_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/clock"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"

	. "bosh-softlayer-cpi/logger"
)

type blockingWriter struct {
	buf bytes.Buffer
	sync.Mutex
}

func (w *blockingWriter) Write(p []byte) (int, error) {
	w.Lock()
	n, err := w.buf.Write(p)
	w.Unlock()
	return n, err
}

func (w *blockingWriter) Len() int {
	w.Lock()
	n := w.buf.Len()
	w.Unlock()
	return n
}

func (w *blockingWriter) String() string {
	w.Lock()
	s := w.buf.String()
	w.Unlock()
	return s
}

var _ = Describe("Logger", func() {
	var (
		logger  Logger
		logBuff bytes.Buffer
		out     blockingWriter
	)
	BeforeEach(func() {
		multiWriter := io.MultiWriter(&out, bufio.NewWriter(&logBuff))

		outLogger := log.New(multiWriter, "", log.LstdFlags)
		errLogger := log.New(&out, "", log.LstdFlags)
		logger = New(boshlog.LevelDebug, "fake-tag-prefix", outLogger, errLogger)
	})

	It("call constructor successfully", func() {
		boshLogger := NewLogger(boshlog.LevelDebug, "fake-tag-prefix")
		boshLogger.Info("LoggerUnitTest", "It is from cpi logger.")
	})

	It("get bosh-utils logger of cpi logger", func() {
		boshLogger := logger.GetBoshLogger()
		boshLogger.Info("LoggerUnitTest", "It is from bosh logger.")
		Expect(out.String()).To(MatchRegexp("\\[LoggerUnitTest\\]"))

		logger.Info("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(MatchRegexp("\\[fake-tag-prefix:LoggerUnitTest\\]"))
	})

	It("get serial tag prefix of cpi logger", func() {
		serialTagPrefix := logger.GetSerialTagPrefix()
		Expect(serialTagPrefix).To(ContainSubstring("fake-tag-prefix"))

	})

	It("get serial tag prefix of cpi logger", func() {
		serialTagPrefix := logger.GetSerialTagPrefix()
		Expect(serialTagPrefix).To(ContainSubstring("fake-tag-prefix"))
	})

	It("call to print DEBUG log", func() {
		logger.Debug("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(ContainSubstring("DEBUG"))
	})

	It("call to print DebugWithDetails log", func() {
		logger.DebugWithDetails("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(ContainSubstring("DEBUG"))
		Expect(out.String()).To(ContainSubstring("********************"))
	})

	It("call to print INFO log", func() {
		logger.Info("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(ContainSubstring("DEBUG"))
	})

	It("call to print WARN log", func() {
		logger.Warn("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(ContainSubstring("WARN"))
	})

	It("call to print ERROR log", func() {
		logger.Error("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(ContainSubstring("ERROR"))
	})

	It("call to print ErrorWithDetails log", func() {
		logger.ErrorWithDetails("LoggerUnitTest", "It is from cpi logger.")
		Expect(out.String()).To(ContainSubstring("ERROR"))
		Expect(out.String()).To(ContainSubstring("********************"))
	})

	It("call ChangeRetryStrategyLogTag", func() {
		execStmtRetryable := boshretry.NewRetryable(
			func() (bool, error) {
				fmt.Println("fake-retry-execute")
				return true, nil
			})
		timeService := clock.NewClock()
		timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(1*time.Minute, 5*time.Second, execStmtRetryable, timeService, logger.GetBoshLogger())
		logger.ChangeRetryStrategyLogTag(&timeoutRetryStrategy)

		err := timeoutRetryStrategy.Try()

		Expect(err).To(BeNil())
		Expect(out.String()).To(ContainSubstring("[fake-tag-prefix:timeoutRetryStrategy] "))
	})
})
