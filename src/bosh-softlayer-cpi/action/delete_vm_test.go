package action_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	fakeaction "github.com/cloudfoundry/bosh-softlayer-cpi/action/fakes"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
)

var _ = Describe("DeleteVM", func() {
	var (
		fakeVmFinder          *fakescommon.FakeVMFinder
		fakeVmDeleterProvider *fakeaction.FakeDeleterProvider
		fakeVmDeleter         *fakescommon.FakeVMDeleter
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVmDeleter = &fakescommon.FakeVMDeleter{}
		fakeVmDeleterProvider = &fakeaction.FakeDeleterProvider{}
	})

	Describe("Run", func() {
		var (
			action      DeleteVMAction
			fakeOptions *ConcreteFactoryOptions

			vmCid VMCID
			err   error
		)

		BeforeEach(func() {
			vmCid = VMCID(1234)
		})

		JustBeforeEach(func() {
			_, err = action.Run(vmCid)
		})
		Context("when delete vm with enable pool succeeds", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: true}},
				}
				action = NewDeleteVM(fakeVmDeleterProvider, *fakeOptions)

				fakeVmDeleterProvider.GetReturns(fakeVmDeleter)
				fakeVmDeleter.DeleteReturns(nil)
			})

			It("fetches deleter by `pool`", func() {
				Expect(fakeVmDeleterProvider.GetCallCount()).To(Equal(1))
				actualKey := fakeVmDeleterProvider.GetArgsForCall(0)
				Expect(actualKey).To(Equal("pool"))
			})

			It("no error return", func() {
				Expect(fakeVmDeleter.DeleteCallCount()).To(Equal(1))
				actualCid := fakeVmDeleter.DeleteArgsForCall(0)
				Expect(actualCid).To(Equal(1234))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when delete vm without enable pool succeeds", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: false}},
				}
				action = NewDeleteVM(fakeVmDeleterProvider, *fakeOptions)

				fakeVmDeleterProvider.GetReturns(fakeVmDeleter)
			})

			It("fetches deleter by `virtualguest`", func() {
				Expect(fakeVmDeleterProvider.GetCallCount()).To(Equal(1))
				actualKey := fakeVmDeleterProvider.GetArgsForCall(0)
				Expect(actualKey).To(Equal("virtualguest"))
			})
		})

		Context("when delete vm error out", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					Softlayer: SoftLayerConfig{FeatureOptions: FeatureOptions{EnablePool: true}},
				}
				action = NewDeleteVM(fakeVmDeleterProvider, *fakeOptions)

				fakeVmDeleterProvider.GetReturns(fakeVmDeleter)
				fakeVmDeleter.DeleteReturns(errors.New("kaboom"))
			})

			It("fetches deleter by `pool`", func() {
				Expect(err.Error()).To(ContainSubstring("kaboom"))
			})
		})
	})
})
