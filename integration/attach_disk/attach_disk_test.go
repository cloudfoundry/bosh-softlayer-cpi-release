package attach_disk_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testhelperscpi "github.com/maximilien/bosh-softlayer-cpi/test_helpers"
	//	util "github.com/maximilien/bosh-softlayer-cpi/util"
	slclient "github.com/maximilien/softlayer-go/client"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	softlayer "github.com/maximilien/softlayer-go/softlayer"
	testhelpers "github.com/maximilien/softlayer-go/test_helpers"
	"log"
)

const configPath = "test_fixtures/cpi_methods/config.json"

var _ = Describe("BOSH Director Level Integration for attach_disk", func() {
	var (
		err error

		client softlayer.Client

		username, apiKey string

		//virtualGuest  datatypes.SoftLayer_Virtual_Guest
		disk          datatypes.SoftLayer_Network_Storage
		createdSshKey datatypes.SoftLayer_Security_Ssh_Key
		//		sshClient     util.SshClient
		vmId int

		accountService      softlayer.SoftLayer_Account_Service
		virtualGuestService softlayer.SoftLayer_Virtual_Guest_Service

		rootTemplatePath, tmpConfigPath string
		strVGID, strDID                 string
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

		replacementMap = map[string]string{
			"Datacenter": testhelpers.GetDatacenter(),
		}

		testhelpers.TIMEOUT = 35 * time.Minute
		testhelpers.POLLING_INTERVAL = 10 * time.Second

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

	Context("attach_disk in SoftLayer with invalid virtual guest id and disk id", func() {
		BeforeEach(func() {
			replacementMap = map[string]string{
				"VMID":   "123456",
				"DiskID": "123",
			}
		})

		It("returns error", func() {
			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("attach_disk", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["error"]).ToNot(BeNil())
		})
	})

	Context("attach_disk in SoftLayer with valid virtual guest id(NO multipath installed) and disk id", func() {
		BeforeEach(func() {
			err = testhelpers.FindAndDeleteTestSshKeys()
			Expect(err).ToNot(HaveOccurred())

			createdSshKey, _ = testhelpers.CreateTestSshKey()
			testhelpers.WaitForCreatedSshKeyToBePresent(createdSshKey.Id)

			//virtualGuest = testhelpers.CreateVirtualGuestAndMarkItTest([]datatypes.SoftLayer_Security_Ssh_Key{createdSshKey})
			createvmJsonPath := filepath.Join(rootTemplatePath, "dev", "create_vm.json")
			f, err := os.Open(createvmJsonPath)
			Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			fb, err := ioutil.ReadAll(f)
			Expect(err).ToNot(HaveOccurred())
			jsonPayload := string(fb)
			log.Println("jsonPayload --> ", jsonPayload)

			log.Println("---> starting create vm")
			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["error"]).To(BeNil())

			id := resultOutput["result"].(string)
			vmId, err = strconv.Atoi(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmId).ToNot(BeNil())
			log.Println("---> created vm ", vmId)
			//

			//			testhelpers.WaitForVirtualGuestToBeRunning(virtualGuest.Id)
			//			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(virtualGuest.Id)
			testhelpers.WaitForVirtualGuestToBeRunning(vmId)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(vmId)

			//vm, err := virtualGuestService.GetObject(virtualGuest.Id)
			vm, err := virtualGuestService.GetObject(vmId)
			Expect(err).ToNot(HaveOccurred())

			disk = testhelpers.CreateDisk(20, strconv.Itoa(vm.Datacenter.Id))

			//strVGID = strconv.Itoa(virtualGuest.Id)
			strVGID = strconv.Itoa(vmId)
			strDID = strconv.Itoa(disk.Id)

			replacementMap = map[string]string{
				"VMID":   strVGID,
				"DiskID": strDID,
			}
		})

		AfterEach(func() {
			//testhelpers.DeleteVirtualGuest(virtualGuest.Id)
			//testhelpers.WaitForVirtualGuestToHaveNoActiveTransactionsOrToErr(virtualGuest.Id)
			testhelpers.DeleteVirtualGuest(vmId)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactionsOrToErr(vmId)
			testhelpers.DeleteDisk(disk.Id)
			testhelpers.DeleteSshKey(createdSshKey.Id)
		})

		It("attach_disk successfully", func() {
			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("attach_disk", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			log.Println("jsonPayload --> ", jsonPayload)

			log.Println("---> starting attach disk")

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["result"]).To(BeNil())
			Expect(resultOutput["error"]).To(BeNil())
		})
	})

		Context("attach_disk in SoftLayer with valid virtual guest id(with multipath installed) and disk id", func() {
		BeforeEach(func() {
			err = testhelpers.FindAndDeleteTestSshKeys()
			Expect(err).ToNot(HaveOccurred())

			createdSshKey, _ = testhelpers.CreateTestSshKey()
			testhelpers.WaitForCreatedSshKeyToBePresent(createdSshKey.Id)

			//virtualGuest = testhelpers.CreateVirtualGuestAndMarkItTest([]datatypes.SoftLayer_Security_Ssh_Key{createdSshKey})
			createvmJsonPath := filepath.Join(rootTemplatePath, "dev", "create_vm.json")
			f, err := os.Open(createvmJsonPath)
			Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			fb, err := ioutil.ReadAll(f)
			Expect(err).ToNot(HaveOccurred())
			jsonPayload := string(fb)
			log.Println("jsonPayload --> ", jsonPayload)

			log.Println("---> starting create vm")
			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["error"]).To(BeNil())

			id := resultOutput["result"].(string)
			vmId, err = strconv.Atoi(id)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmId).ToNot(BeNil())
			log.Println("---> created vm ", vmId)
			//

			//			testhelpers.WaitForVirtualGuestToBeRunning(virtualGuest.Id)
			//			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(virtualGuest.Id)
			testhelpers.WaitForVirtualGuestToBeRunning(vmId)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactions(vmId)

			//vm, err := virtualGuestService.GetObject(virtualGuest.Id)
			vm, err := virtualGuestService.GetObject(vmId)
			Expect(err).ToNot(HaveOccurred())

			disk = testhelpers.CreateDisk(20, strconv.Itoa(vm.Datacenter.Id))

			//strVGID = strconv.Itoa(virtualGuest.Id)
			strVGID = strconv.Itoa(vmId)
			strDID = strconv.Itoa(disk.Id)

			replacementMap = map[string]string{
				"VMID":   strVGID,
				"DiskID": strDID,
			}

			log.Println("---> installing multipath-tools to created vm ", vmId)
			passwords := virtualGuest.OperatingSystem.Passwords
			var rootPassword string
			for _, password := range passwords {
				if password.Username == "root" {
					rootPassword = password.Password
				}
			}

			command := "apt-get install multipath-tools"
			_, err1 := sshClient.ExecCommand("root", rootPassword, virtualGuest.PrimaryBackendIpAddress, command)
			Expect(err1).ToNot(HaveOccurred())
			log.Println("---> multipath-tools installed")
		})

		AfterEach(func() {
			//testhelpers.DeleteVirtualGuest(virtualGuest.Id)
			//testhelpers.WaitForVirtualGuestToHaveNoActiveTransactionsOrToErr(virtualGuest.Id)
			testhelpers.DeleteVirtualGuest(vmId)
			testhelpers.WaitForVirtualGuestToHaveNoActiveTransactionsOrToErr(vmId)
			testhelpers.DeleteDisk(disk.Id)
			testhelpers.DeleteSshKey(createdSshKey.Id)
		})

		It("attach_disk successfully", func() {
			jsonPayload, err := testhelperscpi.GenerateCpiJsonPayload("attach_disk", rootTemplatePath, replacementMap)
			Expect(err).ToNot(HaveOccurred())

			log.Println("jsonPayload --> ", jsonPayload)

			log.Println("---> starting attach disk")

			outputBytes, err := testhelperscpi.RunCpi(rootTemplatePath, tmpConfigPath, jsonPayload)
			log.Println("outputBytes=" + string(outputBytes))
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(outputBytes, &resultOutput)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultOutput["result"]).To(BeNil())
			Expect(resultOutput["error"]).To(BeNil())
		})
	}) 
})
