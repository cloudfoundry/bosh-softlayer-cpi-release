package vm_pool_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/pool"
)

var _ = Describe("DB Util", func() {
	var (
		err    error
		dbPath string
	)

	//TODO: add tests for function: #InitVMPoolDB

	Describe("#OpenDB", func() {
		BeforeEach(func() {
			dbPath, err = ioutil.TempDir("", "OpenDB")
			Expect(err).ToNot(HaveOccurred())

			SQLITE_DB_FILE_PATH = dbPath
		})

		AfterEach(func() {
			os.RemoveAll(dbPath)
		})

		It("returns a new DB", func() {
			db, err := OpenDB(SQLITE_DB_FILE_PATH)
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())
		})

		Context("when SQL_LITE_DB_FILE_PATH is fake", func() {
			BeforeEach(func() {
				SQLITE_DB_FILE_PATH = "fake-sqllite-db-file-path"
			})

			It("fails to return a new DB", func() {
				_, err := OpenDB(SQLITE_DB_FILE_PATH)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("#IsDirectory", func() {
		var path string

		Context("when directory exists", func() {
			BeforeEach(func() {
				path, err = ioutil.TempDir("", "IsDirectory")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll(path)
			})

			It("returns true", func() {
				isDir, err := IsDirectory(path)
				Expect(err).ToNot(HaveOccurred())
				Expect(isDir).To(BeTrue())
			})
		})

		Context("when directory does not exist", func() {
			BeforeEach(func() {
				path = "fake-directory"
			})

			It("returns false", func() {
				isDir, err := IsDirectory(path)
				Expect(err).To(HaveOccurred())
				Expect(isDir).To(BeFalse())
			})
		})

		Context("when directory is a file not a directory", func() {
			BeforeEach(func() {
				file, err := ioutil.TempFile("", "IsDirectory")
				Expect(err).ToNot(HaveOccurred())

				fileInfo, err := file.Stat()
				Expect(err).ToNot(HaveOccurred())

				path = fileInfo.Name()
			})

			AfterEach(func() {
				os.RemoveAll(path)
			})

			It("returns false", func() {
				isDir, err := IsDirectory(path)
				Expect(err).To(HaveOccurred())
				Expect(isDir).To(BeFalse())
			})
		})
	})
})
