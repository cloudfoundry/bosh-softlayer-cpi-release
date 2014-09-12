package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("ConcreteFactoryOptions", func() {
	var (
		options ConcreteFactoryOptions

		validOptions = ConcreteFactoryOptions{
			StemcellsDir: "/tmp/stemcells",

			Agent: bslcvm.AgentOptions{
				Mbus: "fake-mbus",
				NTP:  []string{},

				Blobstore: bslcvm.BlobstoreOptions{
					Type: "fake-blobstore-type",
				},
			},
		}
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			options = validOptions
		})

		It("returns error if StemcellsDir is empty", func() {
			options.StemcellsDir = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty StemcellsDir"))
		})

		It("returns error if agent section is not valid", func() {
			options.Agent.Mbus = ""

			err := options.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Agent configuration"))
		})
	})
})
