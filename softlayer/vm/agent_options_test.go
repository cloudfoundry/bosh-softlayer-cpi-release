package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("AgentOptions", func() {
	var (
		options AgentOptions

		validOptions = AgentOptions{
			Mbus: "fake-mbus",
			NTP:  []string{},

			Blobstore: BlobstoreOptions{
				Type: "fake-blobstore-type",
			},
		}
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			options = validOptions
		})

		It("does not return error if all fields are valid", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if Mbus is empty", func() {
			options.Mbus = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty Mbus"))
		})

		It("returns error if blobstore section is not valid", func() {
			options.Blobstore.Type = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Blobstore configuration"))
		})
	})
})

var _ = Describe("BlobstoreOptions", func() {
	var (
		options BlobstoreOptions

		validOptions = BlobstoreOptions{
			Type:    "fake-type",
			Options: map[string]interface{}{"fake-key": "fake-value"},
		}
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			options = validOptions
		})

		It("does not return error if all fields are valid", func() {
			err := options.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if Type is empty", func() {
			options.Type = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty Type"))
		})
	})
})
