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

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakedisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk/fakes"
	fakestemcell "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
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
		stemcell        *fakestemcell.FakeStemcell
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = fakesutil.NewFakeSshClient()
		agentEnvService = &fakevm.FakeAgentEnvService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		vm = NewSoftLayerVM(1234, softLayerClient, sshClient, agentEnvService, logger, 5*time.Millisecond)

		bslcommon.POLLING_INTERVAL = 1 * time.Millisecond
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
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, 2*time.Millisecond)
				err := vm.DeleteVM()
				Expect(err).ToNot(HaveOccurred())
			})

			It("postCheckActiveTransactionsForDeleteVM time out", func() {
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, 1*time.Millisecond)
				err := vm.DeleteVM()
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
				vm = NewSoftLayerVM(00000, softLayerClient, sshClient, agentEnvService, logger, 2*time.Millisecond)
			})

			It("fails deleting the VM", func() {
				err := vm.DeleteVM()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Reboot", func() {
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("true")
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, 2*time.Millisecond)
			})

			It("reboots the VM successfully", func() {
				err := vm.Reboot()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("invalid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("false")
				vm = NewSoftLayerVM(00000, softLayerClient, sshClient, agentEnvService, logger, 2*time.Millisecond)
			})

			It("fails rebooting the VM", func() {
				err := vm.Reboot()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ReloadOS", func() {
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
					"SoftLayer_Virtual_Guest_Service_reloadOS.json",
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions_NotNone.json",
					"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, 2*time.Millisecond)
				stemcell = fakestemcell.NewFakeStemcell(123456, "5b7bc66a-72c6-447a-94a1-967803fcd76b", "non-dea")
			})

			It("os reload on the VM successfully", func() {
				err := vm.ReloadOS(stemcell, 1*time.Millisecond)
				Expect(err).ToNot(HaveOccurred())
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
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, 2*time.Millisecond)
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

		const expectedDmSetupLs1 = `
36090a0c8600058baa8283574c302c0fc-part1	(252:1)
36090a0c8600058baa8283574c302c0fc	(252:0)
`
		const expectedDmSetupLs2 = `
36090a0c8600058baa8283574c302c0fc-part1	(252:1)
36090a0c8600058baa8283574c302c0fc	(252:0)
36090a0c8600058baa8283574c302c0fd-part1	(252:1)
36090a0c8600058baa8283574c302c0fd	(252:0)
`
		const expectedPartitions1 = `major minor  #blocks  name

   7        0     131072 loop0
 202       16    2097152 xvdb
 202       17    2096451 xvdb1
 202        0   26214400 xvda
 202        1     248832 xvda1
 202        2   25963520 xvda2
 202       32  314572800 xvdc
 202       33  314568733 xvdc1
`
		const expectedPartitions2 = `major minor  #blocks  name

   7        0     131072 loop0
 202       16    2097152 xvdb
 202       17    2096451 xvdb1
 202        0   26214400 xvda
 202        1     248832 xvda1
 202        2   25963520 xvda2
 202       32  314572800 xvdc
 202       33  314568733 xvdc1
 252        0  314572800 dm-0
 252        1  314572799 dm-1
   8       16  314572800 sdb
   8       17  314572799 sdb1
`

		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk(1234)
			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_getObject.json",
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
				"SoftLayer_Network_Storage_Service_getAllowedVirtualGuests_None.json",
				"SoftLayer_Network_Storage_Service_allowAccessFromVirtualGuest.json",
				"SoftLayer_Virtual_Guest_Service_getAllowedHost.json",
				"SoftLayer_Network_Storage_Allowed_Host_Service_getCredential.json",
				"SoftLayer_Virtual_Guest_Service_getUserData_Without_PersistentDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
				"SoftLayer_Virtual_Guest_Service_setMetadata.json",
				"SoftLayer_Virtual_Guest_Service_configureMetadataDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)

		})

		It("attaches the iSCSI volume successfully (multipath-tool installed)", func() {
			expectedCmdResults := []string{
				"/sbin/multipath",
				"No devices found",
				"",
				"",
				"",
				"",
				expectedDmSetupLs1,
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("attaches the iSCSI volume successfully (multipath-tool not installed)", func() {
			expectedCmdResults := []string{
				"",
				expectedPartitions1,
				"",
				"",
				"",
				"",
				expectedPartitions2,
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("attaches second iSCSI volume successfully (multipath-tool installed)", func() {
			expectedCmdResults := []string{
				"/sbin/multipath",
				expectedDmSetupLs1,
				"",
				"",
				"",
				"",
				expectedDmSetupLs2,
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to attach the iSCSI volume", func() {
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, []string{"fake-result"}, errors.New("fake-error"))
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.AttachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DetachDisk", func() {
		var (
			disk bslcdisk.Disk
		)

		const expectedTarget1 = `iqn.2001-05.com.equallogic:0-8a0906-ba580060c-fcc002c3743528a8-fake-user
`
		const expectedTarget2 = `iqn.1992-08.com.netapp:sjc0101
`
		const exportedPortals1 = `10.1.107.49:3260
`
		const expectedPortals2 = `10.1.236.90:3260,1033
10.1.106.75:3260,1030
`
		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk(1234)
			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_getObject.json",
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
				"SoftLayer_Virtual_Guest_Service_getUserData_With_PersistentDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				"SoftLayer_Virtual_Guest_Service_getActiveTransactions_None.json",
				"SoftLayer_Virtual_Guest_Service_setMetadata.json",
				"SoftLayer_Virtual_Guest_Service_configureMetadataDisk.json",
				"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				"SoftLayer_Network_Storage_Service_getAllowedVirtualGuests.json",
				"SoftLayer_Network_Storage_Service_removeAccessFromVirtualGuest.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
		})

		It("detaches legacy iSCSI volume successfully", func() {
			expectedCmdResults := []string{
				expectedTarget1,
				exportedPortals1,
				"",
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("detaches performance storage iSCSI volume successfully", func() {
			expectedCmdResults := []string{
				expectedTarget2,
				expectedPortals2,
				"",
				"",
			}
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, expectedCmdResults, nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to detach iSCSI volume", func() {
			testhelpers.SetTestFixturesForFakeSSHClient(sshClient, []string{"fake-result"}, errors.New("fake-error"))
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger, TIMEOUT_TRANSACTIONS_DELETE_VM)

			err := vm.DetachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})
})
