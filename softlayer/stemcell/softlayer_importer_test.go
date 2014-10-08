package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

var _ = XDescribe("SoftLayerImporter", func() {
	var (
		logger   boshlog.Logger
		importer SoftLayerImporter
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		importer = NewSoftLayerImporter(logger)
	})

	Describe("ImportFromPath", func() {
		It("returns unique stemcell id", func() {
			stemcell, err := importer.ImportFromPath("/fake/path")
			Expect(err).ToNot(HaveOccurred())

			expectedStemcell := NewSoftLayerStemcell("/fake-collection-dir/fake-uuid", logger)
			Expect(stemcell).To(Equal(expectedStemcell))
		})

		It("returns error if generating stemcell id fails", func() {
			stemcell, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-generate-err"))
			Expect(stemcell).To(BeNil())
		})

		It("creates directory in collection directory that will contain unpacked stemcell", func() {
			_, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if creating driectory that will contain unpacked stemcell fails", func() {
			stemcell, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-mkdir-all-err"))
			Expect(stemcell).To(BeNil())
		})

		It("unpacks stemcell into directory that will contain this unpacked stemcell", func() {
			_, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if unpacking stemcell fails", func() {
			_, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).To(HaveOccurred())
		})
	})
})
