package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"
)

var _ = Describe("Info", func() {
	var info Info

	BeforeEach(func() {
		info = NewInfo()
	})

	Describe("Run", func() {
		Context("stemcell_formats", func() {
			var response InfoResult

			BeforeEach(func() {
				var err error

				response, err = info.Run()
				Expect(err).NotTo(HaveOccurred())
			})

			It("supports softlayer-light", func() {
				Expect(response.StemcellFormats).To(ContainElement("softlayer-light"))
			})

		})
	})
})
