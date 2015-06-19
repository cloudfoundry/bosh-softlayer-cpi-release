package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerStemcell", func() {
	var (
		softLayerClient *fakesslclient.FakeSoftLayerClient
		stemcell        SoftLayerStemcell
		logger          boshlog.Logger
	)

	BeforeEach(func() {
		softLayerClient = fakesslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		logger = boshlog.NewLogger(boshlog.LevelNone)

		stemcell = NewSoftLayerStemcell(1234, "fake-stemcell-uuid", DefaultKind, softLayerClient, logger)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			softLayerClient.DoRawHttpRequestResponse = []byte("true")
		})

		Context("when stemcell exist", func() {
			//TODO: GitHub issue #27
			XIt("deletes directory in collection directory that contains unpacked stemcell", func() {
				err := stemcell.Delete()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when stemcell does not exist", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("false")
			})

			It("returns error if deleting stemcell does not exist", func() {
				err := stemcell.Delete()
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
