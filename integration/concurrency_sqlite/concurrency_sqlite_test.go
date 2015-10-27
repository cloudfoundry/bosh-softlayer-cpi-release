package concurrency_sqlite_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/mattn/go-sqlite3"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	common "github.com/maximilien/bosh-softlayer-cpi/common"
	bslcvmpool "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/pool"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

var (
	SQLITE_DB_FOLDER    = "/tmp/concurrency_sqlite_test"
	SQLITE_DB_FILE      = "vm_pool.sqlite"
	SQLITE_DB_FILE_PATH = filepath.Join(SQLITE_DB_FOLDER, SQLITE_DB_FILE)
	stemcellUuid        = "fake_stemcell_uuid"
	domain              = "softalyer.com"
	c                   chan int
	logger              = boshlog.NewLogger(boshlog.LevelInfo)
)

func populateDB() {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 1000; i++ {
		vmID := i
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.InsertVMInfo(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())

	}
}

func insertVMInfo(init int) {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := init; i < init+1000; i++ {
		vmID := i
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		//defer vmInfoDB.CloseDB()

		err = vmInfoDB.InsertVMInfo(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())
	}

	c <- 1
}

func updateVMInfoByID() {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 100; i++ {
		vmID := rand.Intn(100)
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "f", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.UpdateVMInfoByID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())
	}

	c <- 1
}

func queryVMInfobyID() {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 1000; i++ {
		vmID := rand.Intn(1000)
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.QueryVMInfobyID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())

	}

	c <- 1
}

func queryVMInfobyAgentID() {
	db, err := bslcvmpool.OpenDB(SQLITE_DB_FILE_PATH)
	Expect(err).ToNot(HaveOccurred())

	for i := 0; i < 1000; i++ {
		vmID := rand.Intn(1000)
		hostname := fmt.Sprintf("concurrency_test_%d", vmID)
		agentID := strconv.Itoa(i)
		vmInfoDB := bslcvmpool.NewVMInfoDB(vmID, hostname+"."+domain, "t", stemcellUuid, agentID, logger, db)
		defer vmInfoDB.CloseDB()

		err = vmInfoDB.QueryVMInfobyAgentID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		Expect(err).ToNot(HaveOccurred())
	}

	c <- 1
}

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

		It("Manipulate DB concurrently", func() {
			runtime.GOMAXPROCS(2)

			go insertVMInfo(1000)
			go updateVMInfoByID()
			go queryVMInfobyID()
			go queryVMInfobyAgentID()
			go insertVMInfo(2000)
			go updateVMInfoByID()
			go queryVMInfobyID()
			go queryVMInfobyAgentID()
			go insertVMInfo(3000)
			go updateVMInfoByID()
			go queryVMInfobyID()
			go queryVMInfobyAgentID()
			go insertVMInfo(4000)
			go updateVMInfoByID()
			go queryVMInfobyID()
			go queryVMInfobyAgentID()
			go insertVMInfo(5000)
			go updateVMInfoByID()
			go queryVMInfobyID()
			go queryVMInfobyAgentID()

			time.Sleep(60 * time.Second)
		})
	})

})
