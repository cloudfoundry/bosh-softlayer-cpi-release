package delete_stemcell_test

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	testhelperscpi "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	"github.com/cloudfoundry/bosh-utils/logger"
	slclient "github.com/maximilien/softlayer-go/client"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	"github.com/maximilien/softlayer-go/softlayer"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
)

const configPath = "test_fixtures/cpi_methods/config.json"

var _ = Describe("BOSH Director Level Integration for delete_stemcell", func() {
	var (
		err error

		client softlayer.Client

		username, apiKey string

		vgbdtgService softlayer.SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service

		configuration datatypes.SoftLayer_Container_Virtual_Guest_Block_Device_Template_Configuration

		rootTemplatePath, tmpConfigPath string

		virtual_disk_image_id int

		output         map[string]interface{}
		replacementMap map[string]string
	)

	BeforeEach(func() {
		username = os.Getenv("SL_USERNAME")
		Expect(username).ToNot(Equal(""), "username cannot be empty, set SL_USERNAME")

		apiKey = os.Getenv("SL_API_KEY")
		Expect(apiKey).ToNot(Equal(""), "apiKey cannot be empty, set SL_API_KEY")

		client = slclient.NewSoftLayerClient(username, apiKey)
		Expect(client).ToNot(BeNil())

		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		rootTemplatePath = filepath.Join(pwd, "..", "..")

		tmpConfigPath, err = testhelperscpi.CreateTmpConfigPath(rootTemplatePath, configPath, username, apiKey)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err = os.RemoveAll(tmpConfigPath)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("delete_stemcell in SoftLayer", func() {
		BeforeEach(func() {
			swiftUsername := strings.Split(os.Getenv("SWIFT_USERNAME"), ":")[0]
			Expect(swiftUsername).ToNot(Equal(""), "swiftUsername cannot be empty, set SWIFT_USERNAME")

			swiftCluster := os.Getenv("SWIFT_CLUSTER")
			Expect(swiftCluster).ToNot(Equal(""), "swiftCluster cannot be empty, set SWIFT_CLUSTER")

			vgbdtgService, err = client.GetSoftLayer_Virtual_Guest_Block_Device_Template_Group_Service()
			Expect(err).ToNot(HaveOccurred())
			Expect(vgbdtgService).ToNot(BeNil())

			configuration = datatypes.SoftLayer_Container_Virtual_Guest_Block_Device_Template_Configuration{
				Name: "integration-test-vgbtg",
				Note: "",
				OperatingSystemReferenceCode: "UBUNTU_14_64",
				Uri: "swift://" + swiftUsername + "@" + swiftCluster + "/stemcells/bosh-stemcell-4-test.vhd",
			}

			vgbdtGroup, err := vgbdtgService.CreateFromExternalSource(configuration)
			Expect(err).ToNot(HaveOccurred())

			virtual_disk_image_id = vgbdtGroup.Id

			// Wait for transaction to complete
			testhelpers.TIMEOUT = 35 * time.Minute
			testhelpers.POLLING_INTERVAL = 10 * time.Second
			testhelpers.WaitForVirtualGuestBlockTemplateGroupToHaveNoActiveTransactions(vgbdtGroup.Id)
		})

		AfterEach(func() {
			stemcell := bslcstem.NewSoftLayerStemcell(virtual_disk_image_id, "", client, logger.NewLogger(logger.LevelInfo))
			stemcell.Delete()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns nil when passed valid ID", func() {
			replacementMap = map[string]string{
				"ID":         strconv.Itoa(virtual_disk_image_id),
				"Datacenter": testhelpers.GetDatacenter(),
			}
			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("delete_stemcell", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).To(BeNil())
			Expect(output["error"]).To(BeNil())

			time.Sleep(60 * time.Second)
		})
	})

	Context("delete_stemcell in SoftLayer", func() {
		It("returns false because empty parameters", func() {
			jsonPayload := `{"method": "delete_stemcell", "arguments": [],"context": {}}`

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).To(BeNil())
			Expect(output["error"]).ToNot(BeNil())
		})
	})
})
