package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
	. "bosh-softlayer-cpi/softlayer/common"

	fakescommon "bosh-softlayer-cpi/softlayer/common/fakes"
)

var _ = Describe("ConfigureNetworks", func() {
	var (
		fakeVmFinder *fakescommon.FakeVMFinder
		fakeVm       *fakescommon.FakeVM
		action       ConfigureNetworksAction
		networks     Networks
		vmCid        VMCID

		err error
	)

	Describe("Run", func() {
		BeforeEach(func() {
			fakeVmFinder = &fakescommon.FakeVMFinder{}
			fakeVm = &fakescommon.FakeVM{}
			action = NewConfigureNetworks(fakeVmFinder)
			networks = Networks{}

			vmCid = VMCID(123456)
		})

		JustBeforeEach(func() {
			_, err = action.Run(vmCid, networks)
		})

		Context("when configure network succeeds", func() {
			BeforeEach(func() {
				fakeVm.IDReturns(vmCid.Int())
				fakeVmFinder.FindReturns(fakeVm, true, nil)

				fakeVm.ConfigureNetworksReturns(networks, nil)
				fakeVm.ConfigureNetworksSettingsReturns(nil)
			})

			It("fetches vm by cid", func() {
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				actualCid := fakeVmFinder.FindArgsForCall(0)
				Expect(actualCid).To(Equal(123456))
			})

			It("no error return", func() {
				Expect(fakeVm.ConfigureNetworksCallCount()).To(Equal(1))
				actualNetworks := fakeVm.ConfigureNetworksArgsForCall(0)
				Expect(actualNetworks).To(Equal(networks))

				Expect(fakeVm.ConfigureNetworksSettingsCallCount()).To(Equal(1))
			})
		})

		Context("when find vm error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})

		Context("when find vm return false", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(nil, false, nil)
			})

			It("provides relevant error information", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when configure network error out", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.ConfigureNetworksReturns(nil, errors.New("kaboom"))
			})

			It("provides relevant error information", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
})
