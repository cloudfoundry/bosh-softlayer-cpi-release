package vm_test

import (
	"encoding/json"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	sldatatypes "github.com/maximilien/softlayer-go/data_types"
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

	Describe("#AppendPowerDNSToNetworks", func() {
		var (
			networks, expectedNetworks Networks
			cloudProps                 VMCloudProperties
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

			expectedNetworks = map[string]Network{
				"fake-network0": Network{
					Type:    "fake-type",
					IP:      "fake-IP",
					Netmask: "fake-Netmask",
					Gateway: "fake-Gateway",
					DNS: []string{
						"fake-dns0",
						"fake-dns1",
						"fake-powerdns",
					},
					Default:         []string{},
					CloudProperties: map[string]interface{}{},
				},
			}

			cloudProps = VMCloudProperties{
				BoshIp: "fake-powerdns",
			}
		})

		It("returns new networks with PowerDNS", func() {
			pdnsNetworks := AppendPowerDNSToNetworks(networks, cloudProps)
			Expect(pdnsNetworks).ToNot(Equal(Networks{}))
			Expect(pdnsNetworks).To(Equal(expectedNetworks))
		})
	})

	Describe("#Base64EncodeData", func() {
		It("returns base64 encoded string for `fake-data`", func() {
			encodedData := Base64EncodeData("fake-data")
			Expect(encodedData).To(Equal("ZmFrZS1kYXRh"))
		})
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

		BeforeEach(func() {
			agentID = "fake-agentID"
			stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", fakestem.FakeStemcellKind, softLayerClient, logger)
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
					NetworkVlan: sldatatypes.NetworkVlan{Id: 524956}},
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
						Id: 524956,
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
			vgt, err := CreateVirtualGuestTemplate(stemcell, cloudProps)
			Expect(err).ToNot(HaveOccurred())

			//Since VGT.Hostname use timestamp we need to fix it here
			expectedVgt.Hostname = vgt.Hostname
			Expect(vgt).To(Equal(expectedVgt))
		})

	})
})
