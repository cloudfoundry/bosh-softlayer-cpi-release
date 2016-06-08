package set_vm_metadata_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testhelperscpi "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	slclient "github.com/maximilien/softlayer-go/client"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

const configPath = "test_fixtures/cpi_methods/config.json"

var _ = Describe("BOSH Director Level Integration for set_vm_metadata", func() {
	var (
		err error

		client softlayer.Client

		username, apiKey string

		accountService      softlayer.SoftLayer_Account_Service
		virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service

		virtualGuest  datatypes.SoftLayer_Virtual_Guest
		createdSshKey datatypes.SoftLayer_Security_Ssh_Key

		rootTemplatePath, strVGID string

		replacementMap map[string]string
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
	})

	Context("set_vm_metadata", func() {
		BeforeEach(func() {
			err = testhelpers.FindAndDeleteTestSshKeys()
			Expect(err).ToNot(HaveOccurred())

			createdSshKey, _ = testhelpers.CreateTestSshKey()
			testhelpers.WaitForCreatedSshKeyToBePresent(createdSshKey.Id)

			virtualGuest = testhelpers.CreateVirtualGuestAndMarkItTest([]datatypes.SoftLayer_Security_Ssh_Key{createdSshKey})

			testhelpers.WaitForVirtualGuestToBeRunning(virtualGuest.Id)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(virtualGuest.Id)

			pwd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			rootTemplatePath = filepath.Join(pwd, "..", "..")

			strVGID = strconv.Itoa(virtualGuest.Id)

			replacementMap = map[string]string{
				"ID":             strVGID,
				"DirectorUuid":   "fake-director-uuid",
				"Tag_compiling":  "buildpack_python",
				"Tag_deployment": "metadata_deployment",
			}
		})

		AfterEach(func() {
			testhelpers.DeleteVirtualGuest(virtualGuest.Id)
			testhelpers.DeleteSshKey(createdSshKey.Id)
		})

		It("issues set_vm_metadata to cpi", func() {
			pwd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			rootTemplatePath := filepath.Join(pwd, "..", "..")

			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("set_vm_metadata", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			tmpConfigPath, err := testhelperscpi.CreateTmpConfigPath(rootTemplatePath, configPath, username, apiKey)
			Expect(err).ToNot(HaveOccurred())

			_, err = testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			Expect(err).ToNot(HaveOccurred())

			err = os.RemoveAll(tmpConfigPath)
			Expect(err).ToNot(HaveOccurred())

			tagReferences, err := virtualGuestService.GetTagReferences(virtualGuest.Id)
			Expect(err).ToNot(HaveOccurred())

			tagReferencesJSON, err := json.Marshal(tagReferences)
			Expect(err).ToNot(HaveOccurred())

			Ω(tagReferencesJSON).Should(ContainSubstring("buildpack_python"))
			Ω(tagReferencesJSON).Should(ContainSubstring("metadata_deployment"))
		})
	})
})
