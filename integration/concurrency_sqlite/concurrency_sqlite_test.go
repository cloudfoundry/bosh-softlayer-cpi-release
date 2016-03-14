package concurrency_sqlite_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/mattn/go-sqlite3"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	common "github.com/cloudfoundry/bosh-softlayer-cpi/common"
	bslcvmpool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/pool"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

var (
	SQLITE_DB_FOLDER    = "/tmp/concurrency_sqlite_test"
	SQLITE_DB_FILE      = "vm_pool.sqlite"
	SQLITE_DB_FILE_PATH = filepath.Join(SQLITE_DB_FOLDER, SQLITE_DB_FILE)
	stemcellUuid        = "fake_stemcell_uuid"
	domain              = "softlayer.com"
	logger              = boshlog.NewLogger(boshlog.LevelInfo)
)

var _ = Describe("BOSH Director Level Integration for OS Reload", func() {
	var (
		err error
	)

	BeforeEach(func() {
		common.SetOSEnvVariable("OS_RELOAD_ENABLED", "TRUE")
		common.SetOSEnvVariable("SQLITE_DB_FOLDER", SQLITE_DB_FOLDER)
		common.SetOSEnvVariable("SQLITE_DB_FILE", SQLITE_DB_FILE)

		testhelpers.TIMEOUT = 35 * time.Minute
		testhelpers.POLLING_INTERVAL = 10 * time.Second

		err = os.RemoveAll(SQLITE_DB_FOLDER)
		Expect(err).ToNot(HaveOccurred())

		err = bslcvmpool.InitVMPoolDB(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL, logger)
		Expect(err).ToNot(HaveOccurred())

		populateDB()

	})

	AfterEach(func() {
		err = os.RemoveAll(SQLITE_DB_FOLDER)
		Expect(err).ToNot(HaveOccurred())

	})

	Context("Manipulate DB concurrently", func() {

		It("Manipulate DB concurrently", func(done Done) {

			c := make(chan string, 0)

			for i := 200; i <= 1000; i += 200 {
				go insertVMInfo(i, c)
				go updateVMInfoByID(c)
				go queryVMInfobyID(c)
				go queryVMInfobyAgentID(c)
			}

			for i := 0; i < 20; i++ {
				Expect(<-c).To(ContainSubstring("Done!"))
			}

			close(done)

		}, 600)
	})

})

func populateDB() {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 200; i++ {
		vmID := i
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.InsertVMInfo(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())
	}
}

func insertVMInfo(init int, c chan string) {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := init; i < init+200; i++ {
		vmID := i
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.InsertVMInfo(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond)
	}

	c <- "Done!"
}

func updateVMInfoByID(c chan string) {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 200; i++ {
		vmID := rand.Intn(200)
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "f", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.UpdateVMInfoByID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond)
	}

	c <- "Done!"
}

func queryVMInfobyID(c chan string) {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 200; i++ {
		vmID := rand.Intn(200)
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.QueryVMInfobyID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond)
	}

	c <- "Done!"
}

func queryVMInfobyAgentID(c chan string) {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 200; i++ {
		vmID := rand.Intn(200)
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.QueryVMInfobyAgentID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond)
	}

	c <- "Done!"
}
