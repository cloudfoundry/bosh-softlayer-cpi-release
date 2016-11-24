package vm_test

import (
	"encoding/json"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"

	bsldisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"

	fakescommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/fakes"
	fakedisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk/fakes"
	fakestemcell "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell/fakes"
	fakesutil "github.com/cloudfoundry/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	datatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("SoftLayerVirtualGuest", func() {
	var (
		fakeSoftLayerClient *fakeslclient.FakeSoftLayerClient
		sshClient           *fakesutil.FakeSshClient
		agentEnvService     *fakescommon.FakeAgentEnvService
		logger              boshlog.Logger
		vm                  VM
		stemcell            *fakestemcell.FakeStemcell
	)

	BeforeEach(func() {
		fakeSoftLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		sshClient = &fakesutil.FakeSshClient{}
		agentEnvService = &fakescommon.FakeAgentEnvService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		virtualGuest := datatypes.SoftLayer_Virtual_Guest{
			AccountId:                    123456,
			DedicatedAccountHostOnlyFlag: false,
			Domain: "softlayer.com",
			FullyQualifiedDomainName: "fake.softlayer.com",
			Hostname:                 "fake-hostname",
			Id:                       1234567,
			MaxCpu:                   2,
			MaxCpuUnits:              "CORE",
			MaxMemory:                1024,
			StartCpus:                2,
			StatusId:                 1001,
			Uuid:                     "fake-uuid",
			GlobalIdentifier:         "fake-globalIdentifier",
			PrimaryBackendIpAddress:  "fake-primary-backend-ip",
			PrimaryIpAddress:         "fake-primary-ip",
			OperatingSystem: &datatypes.SoftLayer_Operating_System{
				Passwords: []datatypes.SoftLayer_Password{
					datatypes.SoftLayer_Password{
						Username: "fake-root-user",
						Password: "fake-root-password",
					},
				},
			},
		}

		vm = NewSoftLayerVirtualGuest(virtualGuest, fakeSoftLayerClient, sshClient, logger)
		vm.SetAgentEnvService(agentEnvService)
	})

	Describe("Reboot", func() {
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_reboot.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
			})

			It("reboots the VM successfully", func() {
				err := vm.Reboot()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("invalid VM ID is used", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_reboot_fail.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
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
					"SoftLayer_Virtual_Guest_Service_getActiveTransactions.json",
					"SoftLayer_Virtual_Guest_Service_getPowerState.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
				stemcell = &fakestemcell.FakeStemcell{}
			})

			It("os reload on the VM successfully", func() {
				err := vm.ReloadOS(stemcell)
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

				metadata = VMMetadata{}
				err := json.Unmarshal(metadataBytes, &metadata)
				Expect(err).ToNot(HaveOccurred())

				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_setMetadata.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
			})

			It("does not set any tag values on the VM", func() {
				err := vm.SetMetadata(metadata)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("found tags in metadata", func() {
			BeforeEach(func() {
				fileNames := []string{
					"SoftLayer_Virtual_Guest_Service_setMetadata.json",
				}
				testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
			})

			It("at least one tag found", func() {
				metadataBytes := []byte(`{
				  "director": "fake-director-uuid",
				  "deployment": "fake-deployment",
				  "compiling": "buildpack_python"
				}`)

				metadata = VMMetadata{}
				err := json.Unmarshal(metadataBytes, &metadata)
				Expect(err).ToNot(HaveOccurred())

				err = vm.SetMetadata(metadata)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("ConfigureNetworks", func() {
		var (
			networks Networks
		)

		BeforeEach(func() {
			networks = map[string]Network{
				"fake-network0": Network{
					Type:    "fake-type",
					IP:      "fake-IP",
					Netmask: "fake-Netmask",
					Gateway: "fake-Gateway",
					DNS: []string{
						"fake-dns0",
						"fake-dns1",
					},
					Default:         []string{},
					CloudProperties: map[string]interface{}{},
				},
			}
		})

		It("returns the expected network", func() {
			err := vm.ConfigureNetworks(networks)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("#AttachDisk", func() {
		var (
			disk bsldisk.Disk
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
			disk = &fakedisk.FakeDisk{}
			fileNames := []string{
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
				"SoftLayer_Network_Storage_Service_getAllowedVirtualGuests_None.json",
				"SoftLayer_Network_Storage_Service_allowAccessFromVirtualGuest.json",
				"SoftLayer_Virtual_Guest_Service_getAllowedHost.json",
				"SoftLayer_Network_Storage_Allowed_Host_Service_getCredential.json",
				"SoftLayer_Virtual_Guest_Service_getUserData_Without_PersistentDisk.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
		})

		It("attaches the iSCSI volume successfully (multipath-tool installed)", func() {
			expectedCmdResults := []string{
				"/sbin/multipath",
				"No devices found",
				"",
				"",
				"",
				"",
				"",
				expectedDmSetupLs1,
			}
			sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
				return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
			}
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

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
				"",
				expectedPartitions2,
			}
			sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
				return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
			}
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

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
				"",
				expectedDmSetupLs2,
			}

			sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
				return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
			}
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to attach the iSCSI volume", func() {

			sshClient.ExecCommandReturns("fake-result", errors.New("fake-error"))
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

			err := vm.AttachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DetachDisk", func() {
		var (
			disk bsldisk.Disk
		)

		const expectMultipathInstalled = `/sbin/multipath
`
		const expectMountPoints = `/dev/xvda1 on /boot type ext3 (rw,noatime,barrier=0)
rpc_pipefs on /run/rpc_pipefs type rpc_pipefs (rw)
none on /proc/xen type xenfs (rw)
/var/vcap/data/root_tmp on /tmp type ext4 (rw)
/dev/mapper/3600a09803830304f3124457a4575725a-part1 on /var/vcap/store type ext4 (rw)
`

		const expectLogoutIscsi = `Logging out of session [sid: 1, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.67,3260]
Logging out of session [sid: 2, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.52,3260]
Logout of [sid: 1, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.67,3260] successful.
Logout of [sid: 2, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.52,3260] successful.
`
		const expectLoginIscsi = `Logging in to [iface: default, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.67,3260] (multiple)
Logging in to [iface: default, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.52,3260] (multiple)
Login to [iface: default, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.67,3260] successful.
Login to [iface: default, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.52,3260] successful.
`
		const expectStopOpenIscsi = `* Unmounting iscsi-backed filesystems                                                                                                                                                               [ OK ]
 * Disconnecting iSCSI targets                                                                                                                                                                              Logging out of session [sid: 3, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.67,3260]
Logging out of session [sid: 4, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.52,3260]
Logout of [sid: 3, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.67,3260] successful.
Logout of [sid: 4, target: iqn.1992-08.com.netapp:lon0201, portal: 10.1.222.52,3260] successful.
                                                                                                                                                                                                     [ OK ]
 * Stopping iSCSI initiator service
 `
		const expectStartOpenIscsi = `* Starting iSCSI initiator service iscsid                                                                                                                                                           [ OK ]
 * Setting up iSCSI targets
iscsiadm: No records found
 * Mounting network filesystems
 `
		const expectRestartMultipathd = `* Stopping multipath daemon multipathd                                                                                                                                                              [ OK ]
 * Starting multipath daemon multipathd
 `
		BeforeEach(func() {
			disk = &fakedisk.FakeDisk{}
			fileNames := []string{
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
				"SoftLayer_Network_Storage_Service_getAllowedVirtualGuests.json",
				"SoftLayer_Network_Storage_Service_removeAccessFromVirtualGuest.json",
				"SoftLayer_Virtual_Guest_Service_getUserData_With_PersistentDisk.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(fakeSoftLayerClient, fileNames)
		})

		It("detaches iSCSI volume successfully without multipath-tools installed (one volume attached)", func() {
			expectedCmdResults := []string{
				"",
				expectMountPoints,
				"",
				expectStopOpenIscsi,
				"",
				"",
				expectStartOpenIscsi,
			}
			sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
				return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
			}
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("detaches iSCSI volume successfully with multipath-tools installed (one volume attached)", func() {
			expectedCmdResults := []string{
				expectMultipathInstalled,
				"",
				expectMountPoints,
				expectStopOpenIscsi,
				"",
				"",
				expectStartOpenIscsi,
				expectRestartMultipathd,
			}
			sshClient.ExecCommandStub = func(_, _, _, _ string) (string, error) {
				return expectedCmdResults[sshClient.ExecCommandCallCount()-1], nil
			}
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to detach iSCSI volume", func() {
			sshClient.ExecCommandReturns("fake-result", errors.New("fake-error"))
			slh.TIMEOUT = 2 * time.Second
			slh.POLLING_INTERVAL = 1 * time.Second

			err := vm.DetachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})
})
