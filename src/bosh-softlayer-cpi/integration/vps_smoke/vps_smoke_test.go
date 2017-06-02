package vps_smoke_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"log"

	testhelperscpi "bosh-softlayer-cpi/test_helpers"
	"fmt"
	slclient "github.com/maximilien/softlayer-go/client"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

const configPathWithVps = "test_fixtures/cpi_methods/config_vps.json"
const configPathWithoutVps = "test_fixtures/cpi_methods/config.json"

var _ = Describe("BOSH Director Level Integration for create_vm", func() {
	var (
		err error

		client softlayer.Client

		username, apiKey string

		accountService      softlayer.SoftLayer_Account_Service
		virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service

		rootTemplatePath, tmpConfigPath string
		replacementMap                  map[string]string
		resultOutput                    map[string]interface{}
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

		testhelpers.TIMEOUT = 35 * time.Minute
		testhelpers.POLLING_INTERVAL = 10 * time.Second

		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		rootTemplatePath = filepath.Join(pwd, "..", "..")

	})

	AfterEach(func() {
		err = os.RemoveAll(tmpConfigPath)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("create_vm in SoftLayer", func() {
		It("returns valid result because valid parameters", func() {
			replacementMap = map[string]string{
				"Datacenter": testhelpers.GetDatacenter(),
			}

			tmpConfigPath, err = testhelperscpi.CreateTmpConfigPath(rootTemplatePath, configPathWithVps, username, apiKey)
			Expect(err).ToNot(HaveOccurred())

			fmt.Printf("####################: tmpConfigPath: %s \n", tmpConfigPath)

			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("create_vm", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["result"]).ToNot(BeNil())
			Expect(resultOutput["error"]).To(BeNil())

			id := resultOutput["result"].(string)
			vmId, err := strconv.Atoi(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmId).ToNot(BeNil())
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(vmId)
			testhelpers.DeleteVirtualGuest(vmId)

			replacementMap = map[string]string{
				"ID": id,
			}

			tmpConfigPath, err = testhelperscpi.CreateTmpConfigPath(rootTemplatePath, configPathWithoutVps, username, apiKey)
			Expect(err).ToNot(HaveOccurred())

			jsonPayload, err = testhelperscpi.GenerateCpiJsonPayload("delete_vm", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err = testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["error"]).To(BeNil())
		})
	})
})
