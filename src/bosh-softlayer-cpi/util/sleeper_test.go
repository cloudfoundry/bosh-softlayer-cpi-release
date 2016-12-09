package util_test

import (
	"time"

	bscutil "github.com/cloudfoundry/bosh-softlayer-cpi/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sleeper", func() {
	var (
		sleeper bscutil.Sleeper
	)

	Context("RealSleeper", func() {
		BeforeEach(func() {
			sleeper = bscutil.RealSleeper{}
		})

		It("#Sleep", func() {
			startTime := time.Now()
			sleeper.Sleep(1 * time.Second)
			endTime := time.Now()

			Expect(startTime.Before(endTime)).To(BeTrue())
			Expect(endTime.Nanosecond() - startTime.Nanosecond()).To(BeNumerically(">=", 1000))
		})
	})

	Context("RecordingNoopSleeper", func() {
		var recordingNoopSleeper *bscutil.RecordingNoopSleeper

		BeforeEach(func() {
			recordingNoopSleeper = &bscutil.RecordingNoopSleeper{}
			sleeper = recordingNoopSleeper
		})

		Context("#Sleep", func() {
			It("#Sleep", func() {
				recordingNoopSleeper.Sleep(1 * time.Second)

				sleepTimes := recordingNoopSleeper.SleptTimes()
				Expect(len(sleepTimes)).To(Equal(1))
				Expect(sleepTimes).To(ContainElement(1 * time.Second))
			})
		})
	})
})
