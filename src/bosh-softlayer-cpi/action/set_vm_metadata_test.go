package action_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/action"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
)

var _ = Describe("SetVMMetadata", func() {
	var (
		fakeVmFinder *fakescommon.FakeVMFinder
		fakeVm       *fakescommon.FakeVM
		action       SetVMMetadataAction
		metadata     VMMetadata
	)

	BeforeEach(func() {
		fakeVmFinder = &fakescommon.FakeVMFinder{}
		fakeVm = &fakescommon.FakeVM{}
		action = NewSetVMMetadata(fakeVmFinder)

		metadataBytes := []byte(`{
		  "tag1": "dea",
		  "tag2": "test-env",
		  "tag3": "blue"
		}`)

		metadata = VMMetadata{}
		err := json.Unmarshal(metadataBytes, &metadata)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Run", func() {
		var (
			vmCid VMCID
			err   error
		)
		BeforeEach(func() {
			vmCid = VMCID(123456)

		})

		JustBeforeEach(func() {
			_, err = action.Run(vmCid, metadata)
		})
		Context("when set vm metadata succeeds", func() {
			BeforeEach(func() {
				fakeVmFinder.FindReturns(fakeVm, true, nil)
				fakeVm.SetMetadataReturns(nil)
			})

			It("fetches vm by cid", func() {
				Expect(fakeVmFinder.FindCallCount()).To(Equal(1))
				actualCid := fakeVmFinder.FindArgsForCall(0)
				Expect(actualCid).To(Equal(123456))
			})

			It("no error return", func() {
				Expect(fakeVm.SetMetadataCallCount()).To(Equal(1))
				actualMetadata := fakeVm.SetMetadataArgsForCall(0)
				Expect(actualMetadata).To(Equal(metadata))
				Expect(err).NotTo(HaveOccurred())
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
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Finding VM"))
			})
		})
	})
})
