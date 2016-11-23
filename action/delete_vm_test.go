package action_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	fakeaction "github.com/cloudfoundry/bosh-softlayer-cpi/action/fakes"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"

)

var _ = Describe("DeleteVM", func() {
	var (
		fakeVmFinder *fakescommon.FakeVMFinder
		fakeVmDeleterProvider *fakeaction.FakeDeleterProvider
		fakeVmDeleter *fakescommon.FakeVMDeleter
		fakeOptions *ConcreteFactoryOptions

		vmCid VMCID

		action   DeleteVMAction
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVmDeleter = &fakescommon.FakeVMDeleter{}
		fakeVmDeleterProvider = &fakeaction.FakeDeleterProvider{}
		fakeOptions = &ConcreteFactoryOptions{}

		vmCid = VMCID(1234)
		action = NewDeleteVM(fakeVmDeleterProvider, *fakeOptions)
	})

	Describe("Run", func() {
		var (
			err error
		)
		JustBeforeEach(func() {
			_, err = action.Run(vmCid)
		})
		Context("when delete vm with enable pool succeeds", func() {
			BeforeEach(func() {
				fakeOptions = &ConcreteFactoryOptions{
					SoftLayerConfig: SoftLayerConfig{FeatureOptions : FeatureOptions{EnablePool : true}},
				}
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
					SoftLayerConfig: SoftLayerConfig{FeatureOptions : FeatureOptions{EnablePool : false}},
				}
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
					SoftLayerConfig: SoftLayerConfig{FeatureOptions : FeatureOptions{EnablePool : true}},
				}
				fakeVmDeleterProvider.GetReturns(fakeVmDeleter)
				fakeVmDeleter.DeleteReturns(errors.New("kaboom"))
			})

			It("fetches deleter by `pool`", func() {
				Expect(err.Error()).To(ContainSubstring("kaboomr"))
			})
		})
	})
})
