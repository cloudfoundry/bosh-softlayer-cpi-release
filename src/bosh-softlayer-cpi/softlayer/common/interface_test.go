package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/softlayer/common"
	"encoding/json"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

var _ = Describe("Interfaces", func() {
	var (
		vmCloudProps, vmCloudPropsExpect VMCloudProperties
		vmCloudPropsJsonCamelCase        []byte
		vmCloudPropsJsonSnakeCase        []byte
	)

	Describe("#Unmarshal cloudproperties", func() {
		BeforeEach(func() {
			vmCloudPropsExpect = VMCloudProperties{
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
				BoshIp: "1.2.3.4",
				BlockDevices: []sldatatypes.BlockDevice{{
					Device:    "0",
					DiskImage: sldatatypes.DiskImage{Capacity: 100}}},
				NetworkComponents: []sldatatypes.NetworkComponents{{MaxSpeed: 1000}},
			}

			vmCloudPropsJsonCamelCase = []byte(`
			{
			  "vmNamePrefix":"bosh-",
			  "domain":"fake-domain.com",
			  "startCpus":4,
			  "maxMemory":2048,
			  "datacenter":
			  {
			    "name":"fake-datacenter"
			  },
			    "blockDeviceTemplateGroup":
			  {
			    "globalIdentifier":"fake-uuid"
			  },
			  "sshKeys":[
			    {
			      "id":74826
			      }
			  ],
			  "bosh_ip":"1.2.3.4",
			  "rootDiskSize":25,
			  "ephemeralDiskSize":25,
			  "hourlyBillingFlag":true,
			  "localDiskFlag":true,
			  "dedicatedAccountHostOnlyFlag":true,
			  "networkComponents":[
			    {"maxSpeed":1000}
			  ],
			  "primaryNetworkComponent":{
			    "networkVlan":{}
			  },
			  "primaryBackendNetworkComponent":{
			    "networkVlan":{}
			  },
			  "blockDevices":[
			    {
			      "device":"0",
			      "diskImage":{
			      "capacity":100
			      }
			    }
			  ]
			}`)

			vmCloudPropsJsonSnakeCase = []byte(`
			{
			  "vm_name_prefix": "bosh-",
			  "domain": "fake-domain.com",
			  "start_cpus": 4,
			  "max_memory": 2048,
			  "datacenter": {
			    "name": "fake-datacenter"
			  },
			  "block_device_template_group": {
			    "global_identifier": "fake-uuid"
			  },
			  "ssh_keys": [
			    {
			      "id": 74826
			    }
			  ],
			  "bosh_ip":"1.2.3.4",
			  "root_disk_size": 25,
			  "ephemeral_disk_size": 25,
			  "hourly_billing_flag": true,
			  "local_disk_flag": true,
			  "dedicated_account_host_only_flag": true,
			  "network_components": [
			    {
			      "max_speed": 1000
			    }
			  ],
			  "primary_network_component": {
			    "networkVlan": {}
			  },
			  "primary_backend_network_component": {
			    "network_vlan": {}
			  },
			  "block_devices": [
			    {
			      "device": "0",
			      "disk_image": {
				"capacity": 100
			      }
			    }
			  ]
			}`)
		})

		It("return correctly unmarshall snake case VMCloudProperties", func() {
			vmCloudProps = VMCloudProperties{}
			err := json.Unmarshal(vmCloudPropsJsonSnakeCase, &vmCloudProps)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmCloudProps).To(Equal(vmCloudPropsExpect))
		})

		It("return correctly unmarshall camel case VMCloudProperties", func() {
			vmCloudProps = VMCloudProperties{}
			err := json.Unmarshal(vmCloudPropsJsonCamelCase, &vmCloudProps)
			Expect(err).ToNot(HaveOccurred())
			Expect(vmCloudProps).To(Equal(vmCloudPropsExpect))
		})
	})
})
