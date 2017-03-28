package vm_test

import (
	"errors"
	"time"

	fakescommon "bosh-softlayer-cpi/softlayer/common/fakes"
	slh "bosh-softlayer-cpi/softlayer/common/helper"
	testhelpers "bosh-softlayer-cpi/test_helpers"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	. "bosh-softlayer-cpi/softlayer/common"
	. "bosh-softlayer-cpi/softlayer/vm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SoftlayerVirtualGuestDeleter", func() {
	var (
		fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient
		logger              boshlog.Logger
		fakeVmFinder        *fakescommon.FakeVMFinder
		fakeVm              *fakescommon.FakeVM
		deleter             VMDeleter
	)

	BeforeEach(func() {
		fakeSoftLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVm = &fakescommon.FakeVM{}
		deleter = NewSoftLayerVMDeleter(fakeSoftLayerClient, logger, fakeVmFinder)
		slh.TIMEOUT = 2 * time.Second
		slh.POLLING_INTERVAL = 1 * time.Second
	})

	Describe("Delete", func() {
		var (
			err error
		)

		JustBeforeEach(func() {
			err = deleter.Delete(1234567)
			fakeSoftLayerClient.FakeHttpClient.DoRawHttpRequestResponse = []byte("true")
		})

		Context("when deleting virtual guest succeeds", func() {
			BeforeEach(func() {
				fakeVm.IDReturns(1234567)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				setFakeSoftlayerClientDeleteObjectTrueTestFixtures(fakeSoftLayerClient)
			})

			It("returns no error", func() {
				id := fakeVmFinder.FindArgsForCall(0)
				Expect(id).To(Equal(1234567))
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				Expect(fakeVm.DeleteAgentEnvCallCount()).To(Equal(1))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when virtual guest have runing sections", func() {
			BeforeEach(func() {
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("have no pending transactions before deleting vm"))
			})
		})

		Context("when finding virtual guest with id 1234567 fails", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, errors.New("fake-find-error"))
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				Expect(err.Error()).To(ContainSubstring("Finding VirtualGuest with id"))
			})
		})

		Context("when cannot find virtual guest with id 1234567", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, nil)
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				Expect(err.Error()).To(ContainSubstring("Cannot find VirtualGuest with id"))
			})
		})

		Context("when deleting agent env fails", func() {
			BeforeEach(func() {
				fakeVm.IDReturns(1234567)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.DeleteAgentEnvReturns(errors.New("fake-delete-agent-env-error"))
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(fakeVm.DeleteAgentEnvCallCount()).To(Equal(1))
				Expect(err.Error()).To(ContainSubstring("Deleting VM's agent env"))
			})
		})

		Context("when deleting object and error occurs", func() {
			BeforeEach(func() {
				fakeVm.IDReturns(1234567)
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				testhelpers.SetTestFixtureForFakeSoftLayerClient(fakeSoftLayerClient, "SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				Expect(fakeVm.DeleteAgentEnvCallCount()).To(Equal(1))
				Expect(err.Error()).To(ContainSubstring("Deleting SoftLayer VirtualGuest from client"))
			})
		})
	})

})

func setFakeSoftlayerClientDeleteObjectTrueTestFixtures(fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient) {
	fileNames := []string{
		"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
		"SoftLayer_Virtual_Guest_Service_deleteObject_true.json",
	}
	testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
}
