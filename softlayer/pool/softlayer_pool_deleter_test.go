package pool_test

import (
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool"
	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm"
	fakespool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/pool/client/vm/fakes"
	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var _ = Describe("SoftlayerPoolDeleter", func() {
	var (
		softLayerClient         *fakeslclient.FakeSoftLayerClient
		fakeSoftlayerPoolClient *fakespool.FakeSoftLayerPoolClient
		logger                  boshlog.Logger
		deleter                 VMDeleter
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		fakeSoftlayerPoolClient = &fakespool.FakeSoftLayerPoolClient{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		deleter = NewSoftLayerPoolDeleter(fakeSoftlayerPoolClient, softLayerClient, logger)
	})

	Describe("Delete", func() {
		var (
			err error
		)
		JustBeforeEach(func() {
			err = deleter.Delete(1234567)
		})

		Context("when operation vm in pool succeeds", func() {
			BeforeEach(func() {
				fakeSoftlayerPoolClient.GetVMByCidReturns(vm.NewGetVMByCidOK(), nil)
				fakeSoftlayerPoolClient.UpdateVMWithStateReturns(vm.NewUpdateVMWithStateOK(), nil)
			})

			It("get vm by cid", func() {
				Expect(fakeSoftlayerPoolClient.GetVMByCidCallCount()).To(Equal(1))
				getVmByCidParams := fakeSoftlayerPoolClient.GetVMByCidArgsForCall(0)
				Expect(getVmByCidParams.Cid).To(Equal(int32(1234567)))
			})

			It("update vm with state free", func() {
				Expect(fakeSoftlayerPoolClient.UpdateVMWithStateCallCount()).To(Equal(1))
				updateVMWithStateParams := fakeSoftlayerPoolClient.UpdateVMWithStateArgsForCall(0)
				Expect(updateVMWithStateParams.Body.State).To(Equal(models.StateFree))
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when operation vm out of pool succeeds", func() {
			BeforeEach(func() {
				fakeSoftlayerPoolClient.GetVMByCidReturns(nil, vm.NewGetVMByCidNotFound())
				testhelpers.SetTestFixtureForFakeSoftLayerClient(softLayerClient, "SoftLayer_Virtual_Guest_Service_getObject.json")
				fakeSoftlayerPoolClient.AddVMReturns(vm.NewAddVMOK(), nil)
			})

			It("add vm to pool", func() {
				Expect(fakeSoftlayerPoolClient.AddVMCallCount()).To(Equal(1))
				addVmParams := fakeSoftlayerPoolClient.AddVMArgsForCall(0)
				Expect(addVmParams.Body.State).To(Equal(models.StateFree))
				Expect(addVmParams.Body.Cid).To(Equal(int32(1234567)))
			})

			It("no error return", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when get vm by cid from pool error out", func() {
			BeforeEach(func() {
				fakeSoftlayerPoolClient.GetVMByCidReturns(nil, vm.NewGetVMByCidDefault(500))
			})

			It("error return", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Removing vm from pool"))
			})
		})

		Context("when update vm state to free error out", func() {
			BeforeEach(func() {
				fakeSoftlayerPoolClient.GetVMByCidReturns(vm.NewGetVMByCidOK(), nil)
				fakeSoftlayerPoolClient.UpdateVMWithStateReturns(nil, vm.NewUpdateVMWithStateDefault(500))
			})

			It("error return", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Updating state of vm"))
			})
		})
	})
})
