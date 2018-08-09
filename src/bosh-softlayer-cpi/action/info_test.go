package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
)

var _ = Describe("HasVM", func() {
	var (
		action InfoAction
	)

	BeforeEach(func() {
		action = NewInfo()
	})

	Describe("Run", func() {
		Context("stemcell_formats", func() {
			var response InfoResult

			BeforeEach(func() {
				var err error

				response, err = action.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			It("supports softlayer-legacy-light", func() {
				Expect(response.StemcellFormats).To(ContainElement("softlayer-legacy-light"))
			})

		})
	})
})
