package common_test

import (
	. "bosh-softlayer-cpi/softlayer/common"
	"encoding/json"
	"fmt"
	"github.com/maximilien/softlayer-go/softlayer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"time"

	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	testhelpers "bosh-softlayer-cpi/test_helpers"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"

	"errors"
	"os"
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

	Describe("#CreateAgentUserdata", func() {
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
			suffix := fmt.Sprintf("%03d", int(now.UnixNano()/1e6-now.Unix()*1e3))
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
			stemcell = bslcstem.NewSoftLayerStemcell(1234, "fake-stemcell-uuid", softLayerClient, logger)
			cloudProps = VMCloudProperties{
				StartCpus:                    4,
				MaxMemory:                    2048,
				Domain:                       "fake-domain.com",
				EphemeralDiskSize:            25,
				Datacenter:                   "fake-datacenter",
				HourlyBillingFlag:            true,
				LocalDiskFlag:                true,
				VmNamePrefix:                 "bosh-",
				DedicatedAccountHostOnlyFlag: true,
				PrivateNetworkOnlyFlag:       false,
				SshKeys:                      []sldatatypes.SshKey{{Id: 74826}},
				NetworkComponents: []sldatatypes.NetworkComponents{
					{MaxSpeed: 1000},
				},
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
					{MaxSpeed: 1000},
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

				SshKeys: []sldatatypes.SshKey{
					{Id: 74826},
				},

				UserData: []sldatatypes.UserData{
					{Value: "fake-user-data"},
				},
			}
		})

		It("returns a correct virtual guest template", func() {
			vgt, err := CreateVirtualGuestTemplate(stemcell.Uuid(), cloudProps, "fake-user-data", 524954, 524956)
			Expect(err).ToNot(HaveOccurred())

			//Since VGT.Hostname use timestamp we need to fix it here
			expectedVgt.Hostname = vgt.Hostname
			Expect(vgt).To(Equal(expectedVgt))
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

	Describe("#UpdateEtcHostsOfBoshInit", func() {
		const (
			testFilePath1  = "/tmp/test_updateEtcHostsOfBoshInit1"
			testFilePath2  = "/tmp/test_updateEtcHostsOfBoshInit2"
			testFilePath3  = "/test_updateEtcHostsOfBoshInit3"
			appendDNSEntry = "0.0.0.0    bosh-cpi-test.softlayer.com"
		)
		var (
			logger boshlog.Logger
			fs     boshsys.FileSystem
		)

		BeforeEach(func() {
			logger = boshlog.NewWriterLogger(boshlog.LevelError, os.Stderr, os.Stderr)
			fs = boshsys.NewOsFileSystem(logger)
		})
		AfterEach(func() {
			fs.RemoveAll(testFilePath1)
			fs.RemoveAll(testFilePath2)
		})

		Context("when the target file does not exists", func() {
			BeforeEach(func() {
				fs.RemoveAll(testFilePath1)
			})
			It("create the file if the file does not exist", func() {
				err := UpdateEtcHostsOfBoshInit(testFilePath1, appendDNSEntry)
				Expect(err).ToNot(HaveOccurred())
				fileContent, _ := fs.ReadFileString(testFilePath1)
				Ω(fileContent).Should(ContainSubstring(appendDNSEntry))
			})
		})
		Context("when the target file exists", func() {
			BeforeEach(func() {
				fs.RemoveAll(testFilePath2)
				fs.WriteFileString(testFilePath2, "This is the first line")
			})
			It("update file successfully", func() {
				err := UpdateEtcHostsOfBoshInit(testFilePath2, appendDNSEntry)
				Expect(err).ToNot(HaveOccurred())
				fileContent, _ := fs.ReadFileString(testFilePath2)
				Ω(fileContent).Should(ContainSubstring(appendDNSEntry))
			})
		})
		Context("when the target file cannot be created", func() {
			It("returns error due to no permission", func() {
				err := UpdateEtcHostsOfBoshInit(testFilePath3, appendDNSEntry)
				Expect(err).To(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("permission denied"))
			})
		})
	})
})
