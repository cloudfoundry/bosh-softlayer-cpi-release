package vm_pool_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/pool"

	fakes "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/pool/fakes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var _ = Describe("VMInfoDB", func() {
	var (
		vmInfoDB VMInfoDB
		fakeDB   *fakes.FakeDB
		logger   boshlog.Logger
		err      error
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelDebug)

		fakeDB = fakes.NewFakeDB()
	})

	Describe("#CloseDB", func() {
		BeforeEach(func() {
			vmInfoDB = NewVMInfoDB(0, "fake-name", "f", "fake-image-id", "fake-agent-id", logger, fakeDB)
		})

		It("closes the VMInfoDB", func() {
			err = vmInfoDB.CloseDB()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the DB closes with an error", func() {
			BeforeEach(func() {
				fakeDB.CloseError = errors.New("fake close error")
			})

			It("fails to close the VM info DB", func() {
				err = vmInfoDB.CloseDB()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
