package stemcell_test

import (
	"errors"

	boshlog "bosh/logger"
	fakecmd "bosh/platform/commands/fakes"
	fakesys "bosh/system/fakes"
	fakeuuid "bosh/uuid/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
)

var _ = Describe("FSImporter", func() {
	var (
		fs         *fakesys.FakeFileSystem
		uuidGen    *fakeuuid.FakeGenerator
		compressor *fakecmd.FakeCompressor
		logger     boshlog.Logger
		importer   FSImporter
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		uuidGen = &fakeuuid.FakeGenerator{}
		compressor = fakecmd.NewFakeCompressor()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		importer = NewFSImporter("/fake-collection-dir", fs, uuidGen, compressor, logger)
	})

	Describe("ImportFromPath", func() {
		It("returns unique stemcell id", func() {
			uuidGen.GeneratedUuid = "fake-uuid"

			stemcell, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).ToNot(HaveOccurred())

			expectedStemcell := NewFSStemcell("fake-uuid", "/fake-collection-dir/fake-uuid", fs, logger)
			Expect(stemcell).To(Equal(expectedStemcell))
		})

		It("returns error if generating stemcell id fails", func() {
			uuidGen.GenerateError = errors.New("fake-generate-err")

			stemcell, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-generate-err"))
			Expect(stemcell).To(BeNil())
		})

		It("creates directory in collection directory that will contain unpacked stemcell", func() {
			uuidGen.GeneratedUuid = "fake-uuid"

			_, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).ToNot(HaveOccurred())

			unpackDirStat := fs.GetFileTestStat("/fake-collection-dir/fake-uuid")
			Expect(unpackDirStat.FileType).To(Equal(fakesys.FakeFileTypeDir))
			Expect(int(unpackDirStat.FileMode)).To(Equal(0755)) // todo
		})

		It("returns error if creating driectory that will contain unpacked stemcell fails", func() {
			fs.MkdirAllError = errors.New("fake-mkdir-all-err")

			stemcell, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-mkdir-all-err"))
			Expect(stemcell).To(BeNil())
		})

		It("unpacks stemcell into directory that will contain this unpacked stemcell", func() {
			uuidGen.GeneratedUuid = "fake-uuid"

			_, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).ToNot(HaveOccurred())

			Expect(compressor.DecompressFileToDirTarballPaths[0]).To(Equal("/fake-image-path"))
			Expect(compressor.DecompressFileToDirDirs[0]).To(Equal("/fake-collection-dir/fake-uuid"))
		})

		It("returns error if unpacking stemcell fails", func() {
			compressor.DecompressFileToDirErr = errors.New("fake-decompress-error")

			stemcell, err := importer.ImportFromPath("/fake-image-path")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-decompress-err"))
			Expect(stemcell).To(BeNil())
		})
	})
})
