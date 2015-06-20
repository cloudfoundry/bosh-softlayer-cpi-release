package action_test

import (
	"encoding/json"

	. "github.com/maximilien/bosh-softlayer-cpi/action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	action "github.com/maximilien/bosh-softlayer-cpi/action"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
)

var _ = Describe("SetVMMetadata", func() {
	var (
		vmID     action.VMCID
		vmFinder *fakevm.FakeFinder
		action   SetVMMetadata
		metadata bslcvm.VMMetadata
	)

	BeforeEach(func() {
		vmID = 1234
		vmFinder = &fakevm.FakeFinder{}
		action = NewSetVMMetadata(vmFinder)

		metadataBytes := []byte(`{
		  "tag1": "dea",
		  "tag2": "test-env",
		  "tag3": "blue"
		}`)

		metadata = bslcvm.VMMetadata{}
		err := json.Unmarshal(metadataBytes, &metadata)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("#Run", func() {
		Context("when VM could NOT be found", func() {
			BeforeEach(func() {
				vmFinder.FindFound = false
			})

			It("errors with message that VM could not be found", func() {
				_, err := action.Run(vmID, metadata)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when VM could be found", func() {
			BeforeEach(func() {
				vmFinder.FindFound = true
				vmFinder.FindVM = fakevm.NewFakeVM(int(vmID))
			})

			Context("when metadata is not valid", func() {
				Context("when metadata is empty", func() {
					It("does not do anything and returns no error", func() {
						_, err := action.Run(vmID, bslcvm.VMMetadata{})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when metadata is not a hash of string/string", func() {
					BeforeEach(func() {
						metadataBytes := []byte(`{
  							"tag1": 0,
  							"tag2": null,
  							"tag3": false
						}`)

						metadata = bslcvm.VMMetadata{}
						err := json.Unmarshal(metadataBytes, &metadata)
						Expect(err).ToNot(HaveOccurred())
					})

					It("does not do anything and returns no error", func() {
						_, err := action.Run(vmID, metadata)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when metadata is a hash of tags as key/value pairs", func() {
					It("sets each tag on VM", func() {
						_, err := action.Run(vmID, metadata)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})
})
