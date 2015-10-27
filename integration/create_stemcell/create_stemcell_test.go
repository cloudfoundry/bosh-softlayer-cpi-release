package create_stemcell_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testhelperscpi "github.com/maximilien/bosh-softlayer-cpi/test_helpers"
	slclient "github.com/maximilien/softlayer-go/client"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	// testhelpers "github.com/maximilien/softlayer-go/test_helpers"
	"log"
)

const configPath = "test_fixtures/cpi_methods/config.json"

var _ = Describe("BOSH Director Level Integration for create_stemcell", func() {
	var (
		err error

		client softlayer.Client

		username, apiKey string

		vgbdtgService softlayer.SoftLayer_Virtual_Guest_Block_Device_Template_Group_Service

		configuration datatypes.SoftLayer_Container_Virtual_Guest_Block_Device_Template_Configuration

		tmpConfigPath    string
		rootTemplatePath string

		virtual_disk_image_id int

		output map[string]interface{}

		vmId float64
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

	Context("create_stemcell in SoftLayer", func() {
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
				Uri: "swift://" + swiftUsername + "@" + swiftCluster + "/stemcells/test-bosh-stemcell-softlayer.vhd",
			}

			vgbdtGroup, err := vgbdtgService.CreateFromExternalSource(configuration)
			Expect(err).ToNot(HaveOccurred())

			virtual_disk_image_id = vgbdtGroup.Id

			// Wait for transaction to complete
			time.Sleep(1 * time.Minute)
		})

		AfterEach(func() {
			_, err := vgbdtgService.DeleteObject(virtual_disk_image_id)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns true because valid parameters", func() {
			jsonPayload := fmt.Sprintf(
				`{"method": "create_stemcell", "arguments": ["%s", {"virtual-disk-image-id": %d, "virtual-disk-image-uuid": "%s", "datacenter-name": "%s"}],"context": {"director_uuid": "%s"}}`,
				"fake/root/path",
				virtual_disk_image_id,
				"fake-uuid",
				"ams01",
				"fake-director-uuid")
			// jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("create_stemcell", rootTemplatePath, replacementMap)
			// Expect(err).ToNot(HaveOccurred())

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(outputBytes, &output)
			Expect(err).ToNot(HaveOccurred())
			Expect(output["result"]).ToNot(BeNil())
			Expect(output["error"]).To(BeNil())

			vmId = output["result"].(float64)
			Expect(vmId).ToNot(BeZero())
		})
	})

	Context("create_stemcell in SoftLayer", func() {
		It("returns false because empty parameters", func() {
			jsonPayload := `{"method": "create_stemcell", "arguments": [],"context": {}}`

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
