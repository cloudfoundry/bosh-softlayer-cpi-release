package common_test

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"

	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
	"github.com/maximilien/softlayer-go/softlayer"

	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	testhelpers "github.com/cloudfoundry/bosh-softlayer-cpi/test_helpers"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"

	"errors"
	"strings"
)

var _ = Describe("VM Utils", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		logger          boshlog.Logger
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		logger = boshlog.NewLogger(boshlog.LevelNone)
	})

	Describe("#CreateAgentMetadata", func() {
		var (
			agentID, agentName string
			networks           Networks
			disks              DisksSpec
			env                Environment
			agentOptions       AgentOptions
			cloudProps         VMCloudProperties
			expectedMetadata   string
		)

		BeforeEach(func() {
			agentID = "fake-agentID"
			agentName = "fake-agentName"
			cloudProps = VMCloudProperties{
				BoshIp:            "fake-powerdns",
				EphemeralDiskSize: 100,
			}
			networks = Networks{}
			disks = DisksSpec{}
			env = Environment{}
			agentOptions = AgentOptions{}

			expectedMetadata = `{
  "agent_id": "fake-agentID",
  "vm": {
    "name": "vm-fake-agentID",
    "id": "vm-fake-agentID"
  },
  "mbus": "",
  "ntp": null,
  "blobstore": {
    "provider": "",
    "options": null
  },
  "networks": {

  },
  "disks": {
    "ephemeral": "/dev/xvdc",
    "persistent": null
  },
  "env": {

  }
}`
		})

		It("return agent metadata", func() {
			userdata := CreateAgentUserData(agentID, cloudProps, networks, env, agentOptions)
			Expect(json.Marshal(userdata)).To(MatchJSON(expectedMetadata))
		})
	})

	Describe("#TimeStampForTime", func() {
		It("returns a formatted timestamp for time value", func() {
			now := time.Now()
			timeStamp := TimeStampForTime(now)
			Expect(timeStamp).ToNot(Equal(""))
			prefix := now.Format("20060102-030405-")
			suffix := strconv.Itoa(int(now.UnixNano()/1e6 - now.Unix()*1e3))
			Expect(timeStamp).To(Equal(prefix + suffix))
		})
	})

	Describe("#CreateDisksSpec", func() {
		Context("when ephemeral disk size is 0", func() {
			It("returns an empty DisksSpec", func() {
				disksSpec := CreateDisksSpec(0)
				Expect(disksSpec).To(Equal(DisksSpec{}))
			})
		})

		Context("when ephemeral disk size is > 0", func() {
			var expectedDisksSpec DisksSpec
			BeforeEach(func() {
				expectedDisksSpec = DisksSpec{
					Ephemeral:  "/dev/xvdc",
					Persistent: nil,
				}
			})

			It("returns a DisksSpec with device named /dev/xvdc", func() {
				disksSpec := CreateDisksSpec(1)
				Expect(disksSpec).To(Equal(expectedDisksSpec))
			})
		})
	})

	Describe("#CreateVirtualGuestTemplate", func() {
		var (
			agentID      string
			stemcell     bslcstem.SoftLayerStemcell
			cloudProps   VMCloudProperties
			networks     Networks
			env          Environment
			agentOptions AgentOptions
			expectedVgt  sldatatypes.SoftLayer_Virtual_Guest_Template
		)

		Context("when PrimaryNetworkComponent, PrimaryBackendNetworkComponent exist in cloudProps", func() {
			BeforeEach(func() {
				agentID = "fake-agentID"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
				cloudProps = VMCloudProperties{
					StartCpus: 4,
					MaxMemory: 2048,
					Domain:    "fake-domain.com",
					BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-uuid",
					},
					RootDiskSize:                 25,
					EphemeralDiskSize:            25,
					Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					VmNamePrefix:                 "bosh-",
					PostInstallScriptUri:         "",
					DedicatedAccountHostOnlyFlag: true,
					PrivateNetworkOnlyFlag:       false,
					SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
					BlockDevices: []sldatatypes.BlockDevice{{
						Device:    "0",
						DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
					NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
					PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{Id: 524954}},
					PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
				}

				networks = Networks{}
				env = Environment{}
				agentOptions = AgentOptions{}

				expectedVgt = sldatatypes.SoftLayer_Virtual_Guest_Template{
					Hostname:  "bosh-20150810-081217-541",
					Domain:    "fake-domain.com",
					StartCpus: 4,
					MaxMemory: 2048,

					Datacenter: sldatatypes.Datacenter{
						Name: "fake-datacenter",
					},

					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					OperatingSystemReferenceCode: "",

					BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-stemcell-uuid",
					},

					DedicatedAccountHostOnlyFlag: true,

					NetworkComponents: []sldatatypes.NetworkComponents{
						sldatatypes.NetworkComponents{MaxSpeed: 1000},
					},

					PrivateNetworkOnlyFlag: false,

					PrimaryNetworkComponent: &sldatatypes.PrimaryNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 524954,
						},
					},

					PrimaryBackendNetworkComponent: &sldatatypes.PrimaryBackendNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 524956,
						},
					},

					BlockDevices: []sldatatypes.BlockDevice{
						sldatatypes.BlockDevice{
							Device:    "0",
							DiskImage: sldatatypes.DiskImage{Capacity: 100},
						},
					},

					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 74826},
					},

					PostInstallScriptUri: "",
				}
			})

			It("returns a correct virtual guest template", func() {
				vgt, err := CreateVirtualGuestTemplate(stemcell, cloudProps, networks)
				Expect(err).ToNot(HaveOccurred())

				//Since VGT.Hostname use timestamp we need to fix it here
				expectedVgt.Hostname = vgt.Hostname
				Expect(vgt).To(Equal(expectedVgt))
			})
		})

		Context("when PrimaryNetworkComponent, PrimaryBackendNetworkComponent exist in network settings", func() {
			BeforeEach(func() {
				agentID = "fake-agentID"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
				cloudProps = VMCloudProperties{
					StartCpus: 4,
					MaxMemory: 2048,
					Domain:    "fake-domain.com",
					BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-uuid",
					},
					RootDiskSize:                 25,
					EphemeralDiskSize:            25,
					Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					VmNamePrefix:                 "bosh-",
					PostInstallScriptUri:         "",
					DedicatedAccountHostOnlyFlag: true,
					SshKeys: []sldatatypes.SshKey{{Id: 74826}},
					BlockDevices: []sldatatypes.BlockDevice{{
						Device:    "0",
						DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
					NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
				}

				networks = Networks{
					"fake-net-name": Network{
						Type: "dynamic",

						IP:      "fake-ip",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",

						DNS:           []string{"fake-dns"},
						Default:       []string{"fake-default"},
						Preconfigured: true,

						CloudProperties: map[string]interface{}{
							"PrimaryNetworkComponent": map[string]interface{}{
								"NetworkVlan": map[string]interface{}{
									"Id": float64(524954),
								},
							},
							"PrimaryBackendNetworkComponent": map[string]interface{}{
								"NetworkVlan": map[string]interface{}{
									"Id": float64(524956),
								},
							},
						},
					},
				}

				env = Environment{}
				agentOptions = AgentOptions{}

				expectedVgt = sldatatypes.SoftLayer_Virtual_Guest_Template{
					Hostname:  "bosh-20150810-081217-541",
					Domain:    "fake-domain.com",
					StartCpus: 4,
					MaxMemory: 2048,

					Datacenter: sldatatypes.Datacenter{
						Name: "fake-datacenter",
					},

					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					OperatingSystemReferenceCode: "",

					BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-stemcell-uuid",
					},

					DedicatedAccountHostOnlyFlag: true,

					NetworkComponents: []sldatatypes.NetworkComponents{
						sldatatypes.NetworkComponents{MaxSpeed: 1000},
					},

					PrivateNetworkOnlyFlag: false,

					PrimaryNetworkComponent: &sldatatypes.PrimaryNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 524954,
						},
					},

					PrimaryBackendNetworkComponent: &sldatatypes.PrimaryBackendNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 524956,
						},
					},

					BlockDevices: []sldatatypes.BlockDevice{
						sldatatypes.BlockDevice{
							Device:    "0",
							DiskImage: sldatatypes.DiskImage{Capacity: 100},
						},
					},

					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 74826},
					},

					PostInstallScriptUri: "",
				}
			})

			It("returns a correct virtual guest template", func() {
				vgt, err := CreateVirtualGuestTemplate(stemcell, cloudProps, networks)
				Expect(err).ToNot(HaveOccurred())

				//Since VGT.Hostname use timestamp we need to fix it here
				expectedVgt.Hostname = vgt.Hostname
				Expect(vgt).To(Equal(expectedVgt))
			})
		})

		Context("when PrimaryBackendNetworkComponent, PrivateNetworkOnlyFlag exist in network settings", func() {
			BeforeEach(func() {
				agentID = "fake-agentID"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
				cloudProps = VMCloudProperties{
					StartCpus: 4,
					MaxMemory: 2048,
					Domain:    "fake-domain.com",
					BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-uuid",
					},
					RootDiskSize:                 25,
					EphemeralDiskSize:            25,
					Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					VmNamePrefix:                 "bosh-",
					PostInstallScriptUri:         "",
					DedicatedAccountHostOnlyFlag: true,
					SshKeys: []sldatatypes.SshKey{{Id: 74826}},
					BlockDevices: []sldatatypes.BlockDevice{{
						Device:    "0",
						DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
					NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
				}

				networks = Networks{
					"fake-net-name": Network{
						Type: "dynamic",

						IP:      "fake-ip",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",

						DNS:           []string{"fake-dns"},
						Default:       []string{"fake-default"},
						Preconfigured: true,

						CloudProperties: map[string]interface{}{
							"PrimaryBackendNetworkComponent": map[string]interface{}{
								"NetworkVlan": map[string]interface{}{
									"Id": float64(524956),
								},
							},
							"PrivateNetworkOnlyFlag": true,
						},
					},
				}

				env = Environment{}
				agentOptions = AgentOptions{}

				expectedVgt = sldatatypes.SoftLayer_Virtual_Guest_Template{
					Hostname:  "bosh-20150810-081217-541",
					Domain:    "fake-domain.com",
					StartCpus: 4,
					MaxMemory: 2048,

					Datacenter: sldatatypes.Datacenter{
						Name: "fake-datacenter",
					},

					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					OperatingSystemReferenceCode: "",

					BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-stemcell-uuid",
					},

					DedicatedAccountHostOnlyFlag: true,

					NetworkComponents: []sldatatypes.NetworkComponents{
						sldatatypes.NetworkComponents{MaxSpeed: 1000},
					},

					PrivateNetworkOnlyFlag: true,

					PrimaryNetworkComponent: &sldatatypes.PrimaryNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 0,
						},
					},
					PrimaryBackendNetworkComponent: &sldatatypes.PrimaryBackendNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 524956,
						},
					},

					BlockDevices: []sldatatypes.BlockDevice{
						sldatatypes.BlockDevice{
							Device:    "0",
							DiskImage: sldatatypes.DiskImage{Capacity: 100},
						},
					},

					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 74826},
					},

					PostInstallScriptUri: "",
				}
			})

			It("returns a correct virtual guest template", func() {
				vgt, err := CreateVirtualGuestTemplate(stemcell, cloudProps, networks)
				Expect(err).ToNot(HaveOccurred())

				//Since VGT.Hostname use timestamp we need to fix it here
				expectedVgt.Hostname = vgt.Hostname
				Expect(vgt).To(Equal(expectedVgt))
			})
		})

		Context("when PrimaryBackendNetworkComponent, PrimaryBackendNetworkComponent exist in both cloudProps and network settings", func() {
			BeforeEach(func() {
				agentID = "fake-agentID"
				stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
				cloudProps = VMCloudProperties{
					StartCpus: 4,
					MaxMemory: 2048,
					Domain:    "fake-domain.com",
					BlockDeviceTemplateGroup: sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-uuid",
					},
					RootDiskSize:                 25,
					EphemeralDiskSize:            25,
					Datacenter:                   sldatatypes.Datacenter{Name: "fake-datacenter"},
					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					VmNamePrefix:                 "bosh-",
					PostInstallScriptUri:         "",
					DedicatedAccountHostOnlyFlag: true,
					PrivateNetworkOnlyFlag:       false,
					SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
					BlockDevices: []sldatatypes.BlockDevice{{
						Device:    "0",
						DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
					NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
					PrimaryNetworkComponent: sldatatypes.PrimaryNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{Id: 524954}},
					PrimaryBackendNetworkComponent: sldatatypes.PrimaryBackendNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
				}

				networks = Networks{
					"fake-net-name": Network{
						Type: "dynamic",

						IP:      "fake-ip",
						Netmask: "fake-netmask",
						Gateway: "fake-gateway",

						DNS:           []string{"fake-dns"},
						Default:       []string{"fake-default"},
						Preconfigured: true,

						CloudProperties: map[string]interface{}{
							"PrimaryNetworkComponent": map[string]interface{}{
								"NetworkVlan": map[string]interface{}{
									"Id": float64(123456),
								},
							},
							"PrimaryBackendNetworkComponent": map[string]interface{}{
								"NetworkVlan": map[string]interface{}{
									"Id": float64(123456),
								},
							},
						},
					},
				}

				env = Environment{}
				agentOptions = AgentOptions{}

				expectedVgt = sldatatypes.SoftLayer_Virtual_Guest_Template{
					Hostname:  "bosh-20150810-081217-541",
					Domain:    "fake-domain.com",
					StartCpus: 4,
					MaxMemory: 2048,

					Datacenter: sldatatypes.Datacenter{
						Name: "fake-datacenter",
					},

					HourlyBillingFlag:            true,
					LocalDiskFlag:                true,
					OperatingSystemReferenceCode: "",

					BlockDeviceTemplateGroup: &sldatatypes.BlockDeviceTemplateGroup{
						GlobalIdentifier: "fake-stemcell-uuid",
					},

					DedicatedAccountHostOnlyFlag: true,

					NetworkComponents: []sldatatypes.NetworkComponents{
						sldatatypes.NetworkComponents{MaxSpeed: 1000},
					},

					PrivateNetworkOnlyFlag: false,

					PrimaryNetworkComponent: &sldatatypes.PrimaryNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 123456,
						},
					},

					PrimaryBackendNetworkComponent: &sldatatypes.PrimaryBackendNetworkComponent{
						NetworkVlan: sldatatypes.NetworkVlan{
							Id: 123456,
						},
					},

					BlockDevices: []sldatatypes.BlockDevice{
						sldatatypes.BlockDevice{
							Device:    "0",
							DiskImage: sldatatypes.DiskImage{Capacity: 100},
						},
					},

					SshKeys: []sldatatypes.SshKey{
						sldatatypes.SshKey{Id: 74826},
					},

					PostInstallScriptUri: "",
				}
			})

			It("returns a correct virtual guest template", func() {
				vgt, err := CreateVirtualGuestTemplate(stemcell, cloudProps, networks)
				Expect(err).ToNot(HaveOccurred())

				//Since VGT.Hostname use timestamp we need to fix it here
				expectedVgt.Hostname = vgt.Hostname
				Expect(vgt).To(Equal(expectedVgt))
			})
		})
	})

	Describe("#UpdateDeviceName", func() {
		var (
			cloudProps VMCloudProperties
			slvgs      softlayer.SoftLayer_Virtual_Guest_Service
		)
		BeforeEach(func() {
			cloudProps = VMCloudProperties{
				VmNamePrefix: "fake-hostname",
				Domain:       "fake-domain",
			}
			slvgs, _ = softLayerClient.GetSoftLayer_Virtual_Guest_Service()

			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_editObject.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
		})
		Context("when device name is updated successfully", func() {
			It("returns nil from EditObject", func() {
				err := UpdateDeviceName(1, slvgs, VMCloudProperties{})
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("when fails to update device name", func() {
			BeforeEach(func() {
				softLayerClient.FakeHttpClient.DoRawHttpRequestError = errors.New("Error occurred when updating the object")
			})
			It("returns error from EditObject", func() {
				err := UpdateDeviceName(1, slvgs, VMCloudProperties{})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("#GetLocalIPAddressOfGivenInterface", func() {
		var (
			ip         string
			interfaces []net.Interface
			err        error
		)
		BeforeEach(func() {
			interfaces, _ = net.Interfaces()
		})
		Context("when specifying a valid network interface", func() {
			It("returns the correct IPv4 address", func() {
				for _, i := range interfaces {
					if i.Name == "eth0" || i.Name == "en0" {
						ip, err = GetLocalIPAddressOfGivenInterface(i.Name)
						break
					}
				}

				Expect(err).ToNot(HaveOccurred())
				Expect(3).To(Equal(strings.Count(ip, ".")))
			})
		})
		Context("when specifying an invalid network interface", func() {
			It("returns an empty string and error", func() {
				ip, err = GetLocalIPAddressOfGivenInterface("wrongInterface")

				Expect("").To(Equal(ip))
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
