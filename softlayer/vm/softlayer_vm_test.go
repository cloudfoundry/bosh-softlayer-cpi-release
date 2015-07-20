package vm_test

import (
	"encoding/json"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakedisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakesutil "github.com/maximilien/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerVM", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		sshClient       *fakesutil.FakeSshClient
		agentEnvService *fakevm.FakeAgentEnvService
		logger          boshlog.Logger
		vm              SoftLayerVM
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		agentEnvService = &fakevm.FakeAgentEnvService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		vm = NewSoftLayerVM(1234, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
		})
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_deleteObject_true.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_NotNone.json",
					"SoftLayer_Virtual_Guest_Service_getObject.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransaction.json",
					"SoftLayer_Virtual_Guest_Service_getEmptyObject.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
			})

			It("deletes the VM successfully", func() {
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)
				err := vm.Delete()
				Expect(err).ToNot(HaveOccurred())
			})

			It("postCheckActiveTransactionsForDeleteVM time out", func() {
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, 5*time.Second)
				err := vm.Delete()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("invalid VM ID is used", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_NotNone.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_deleteObject_false.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_NotNone.json",
					"SoftLayer_Virtual_Guest_Service_getObject.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransaction.json",
					"SoftLayer_Virtual_Guest_Service_getEmptyObject.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
				vm = NewSoftLayerVM(00000, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)
			})

			It("fails deleting the VM", func() {
				err := vm.Delete()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Reboot", func() {
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("true")
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)
			})

			It("reboots the VM successfully", func() {
				err := vm.Reboot()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("invalid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("false")
				vm = NewSoftLayerVM(00000, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)
			})

			It("fails rebooting the VM", func() {
				err := vm.Reboot()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("SetMetadata", func() {
		var (
			metadata VMMetadata
		)

		Context("no tags found in metadata", func() {
			BeforeEach(func() {
				metadataBytes := []byte(`{
				  "director": "fake-director-uuid",
				  "name": "fake-director"
				}`)

				metadata = bslcvm.VMMetadata{}
				err := json.Unmarshal(metadataBytes, &metadata)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not set any tag values on the VM", func() {
				err := vm.SetMetadata(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(softLayerClient.DoRawHttpRequestResponseCount).To(Equal(0))
			})
		})

		Context("found tags in metadata", func() {
			BeforeEach(func() {
				metadataBytes := []byte(`{
				  "director": "fake-director-uuid",
				  "name": "fake-director",
				  "tags": "test, tag, director"
				}`)

				metadata = bslcvm.VMMetadata{}
				err := json.Unmarshal(metadataBytes, &metadata)
				Expect(err).ToNot(HaveOccurred())

				softLayerClient.DoRawHttpRequestResponse = []byte("true")
			})

			It("the tags value is empty", func() {
				metadata["tags"] = ""
				err := vm.SetMetadata(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(softLayerClient.DoRawHttpRequestResponseCount).To(Equal(0))
			})

			It("at least one tag found", func() {
				err := vm.SetMetadata(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(softLayerClient.DoRawHttpRequestResponseCount).To(Equal(1))
			})

			Context("when SLVG.SetTags call fails", func() {
				BeforeEach(func() {
					softLayerClient.DoRawHttpRequestError = errors.New("fake-error")
				})

				It("fails with error", func() {
					err := vm.SetMetadata(metadata)

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("ConfigureNetworks", func() {
		var (
			networks Networks
		)

		BeforeEach(func() {
			networks = Networks{}
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)
		})

		It("returns NotSupportedError", func() {
			err := vm.ConfigureNetworks(networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Not supported"))
			Expect(err.(NotSupportedError).Type()).To(Equal("Bosh::Clouds::NotSupported"))
		})
	})

	Describe("#AttachDisk", func() {
		var (
			disk bslcdisk.Disk
		)
		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk(1234)
			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_getObject.json",
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
				"SoftLayer_Network_Storage_Service_getAllowedVirtualGuests_None.json",
				"SoftLayer_Network_Storage_Service_allowAccessFromVirtualGuest.json",
				"SoftLayer_Virtual_Guest_Service_getUserData_Without_PersistentDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
				"SoftLayer_Virtual_Guest_Service_setMetadata.json",
				"SoftLayer_Virtual_Guest_Service_configureMetadataDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
		})

		It("attaches the iSCSI volume successfully", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to attach the iSCSI volume", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", errors.New("fake-error"))
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.AttachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DetachDisk", func() {
		var (
			disk bslcdisk.Disk
		)
		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk(1234)
			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_getObject.json",
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
				"SoftLayer_Network_Storage_Service_getAllowedVirtualGuests.json",
				"SoftLayer_Network_Storage_Service_removeAccessFromVirtualGuest.json",
				"SoftLayer_Virtual_Guest_Service_getUserData_With_PersistentDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
				"SoftLayer_Virtual_Guest_Service_setMetadata.json",
				"SoftLayer_Virtual_Guest_Service_configureMetadataDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
		})

		It("detaches the iSCSI volume successfully", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to detach the iSCSI volume", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", errors.New("fake-error"))
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.DetachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})
})
