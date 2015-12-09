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
		err                 error
		dbPath              string
		SQLITE_DB_FILE_PATH string
	)

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
			db, err := OpenDB(dbPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(db).ToNot(BeNil())
		})
	})

})
