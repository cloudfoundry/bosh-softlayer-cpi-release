package os_reload_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"log"

	testhelperscpi "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	slclient "github.com/maximilien/softlayer-go/client"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

const configPath = "test_fixtures/cpi_methods/config.json"

var _ = Describe("BOSH Director Level Integration for OS Reload", func() {
	var (
		err error

		client softlayer.Client

		username, apiKey string

		accountService      softlayer.SoftLayer_Account_Service
		virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service

		rootTemplatePath, tmpConfigPath string
		replacementMap                  map[string]string

		output map[string]interface{}

		vmId int
	)

	BeforeEach(func() {
		username = os.Getenv("SL_USERNAME")
		Expect(username).ToNot(Equal(""), "username cannot be empty, set SL_USERNAME")

		apiKey = os.Getenv("SL_API_KEY")
		Expect(apiKey).ToNot(Equal(""), "apiKey cannot be empty, set SL_API_KEY")

		client = slclient.NewSoftLayerClient(username, apiKey)
		Expect(client).ToNot(BeNil())

		accountService, err = testhelpers.CreateAccountService()
		Expect(err).ToNot(HaveOccurred())

		virtualGuestService, err = testhelpers.CreateVirtualGuestService()
		Expect(err).ToNot(HaveOccurred())

		os.Setenv("SQLITE_DB_FOLDER", "/tmp")

		testhelpers.TIMEOUT = 35 * time.Minute
		testhelpers.POLLING_INTERVAL = 10 * time.Second

		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		rootTemplatePath = filepath.Join(pwd, "..", "..")

		tmpConfigPath, err = testhelperscpi.CreateTmpConfigPath(rootTemplatePath, configPath, username, apiKey)
		Expect(err).ToNot(HaveOccurred())

	})

	AfterEach(func() {
		testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(int(vmId))
		testhelpers.DeleteVirtualGuest(int(vmId))

		err = os.RemoveAll(tmpConfigPath)
		Expect(err).ToNot(HaveOccurred())

	})

	Context("True path of OS Reload", func() {

		BeforeEach(func() {
			os.Setenv("SQLITE_DB_FILE", "vm_pool.sqlite.right")
		})

		AfterEach(func() {
			err = os.Remove("/tmp/vm_pool.sqlite.right")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns true if the VM creation procedure is correct", func() {

			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("create_vm", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes of vm creation: " + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).ToNot(BeNil())
			Expect(output["error"]).To(BeNil())

			id := output["result"].(string)
			vmId, err = strconv.Atoi(id)
			log.Println(fmt.Sprintf("Create a new VM with ID: %d", vmId))

			strVGID := id
			replacementMap1 := map[string]string{
				"ID":           strVGID,
				"DirectorUuid": "fake-director-uuid",
			}

			jsonPayload, err = testhelperscpi.GenerateCpiJsonPayload("delete_vm", rootTemplatePath, replacementMap1)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err = testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes of vm deletion: " + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)

			Expect(err).ToNot(HaveOccurred())
			Expect(output["error"]).To(BeNil())

			log.Println(fmt.Sprintf("Delete the VM virtually with ID: %d", int(vmId)))

			jsonPayload, err = testhelperscpi.GenerateCpiJsonPayload("create_vm", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err = testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).ToNot(BeNil())
			Expect(output["error"]).To(BeNil())

			id = output["result"].(string)
			vmId, err = strconv.Atoi(id)
			Expect(vmId).ToNot(BeZero())
			log.Println(fmt.Sprintf("OS reload on VM with ID: %d", int(vmId)))
		})
	})

	Context("False path of OS Reload", func() {

		BeforeEach(func() {
			os.Setenv("SQLITE_DB_FILE", "vm_pool.sqlite.wrong")
		})

		AfterEach(func() {
			err = os.Remove("/tmp/vm_pool.sqlite.wrong")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns false if the VM creation procedure is wrong", func() {
			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("create_vm", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).ToNot(BeNil())
			Expect(output["error"]).To(BeNil())

			id := output["result"].(string)
			vmId, err = strconv.Atoi(id)
			Expect(vmId).ToNot(BeZero())
			log.Println(fmt.Sprintf("Create a new VM with ID: %d", vmId))

			jsonPayload, err = testhelperscpi.GenerateCpiJsonPayload("create_vm", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err = testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).To(BeNil())
			Expect(output["error"]).ToNot(BeNil())

		})
	})

})
