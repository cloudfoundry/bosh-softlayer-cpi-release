package integration

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
)

var _ = Describe("VM", func() {
	It("creates a VM with an invalid configuration and receives an error message", func() {
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "datdcenter-error"
		    },
		    {
		      "default": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		resp, err := execCPI(request)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error.Message).ToNot(BeEmpty())
	})

	It("executes the VM lifecycle", func() {
		var vmCID string
		By("creating a VM")
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "default": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)

		By("locating the VM")
		request = fmt.Sprintf(`{
		  "method": "has_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		exists := assertSucceedsWithResult(request).(bool)
		Expect(exists).To(Equal(true))

		By("Setting the metadata")
		request = fmt.Sprintf(`{
		  "method": "set_vm_metadata",
		  "arguments": [
		    "%v",
		    {
		      "cpi": "softlayer-cpi",
		      "test-job": "integration"
		    }
		  ]
		}`, vmCID)
		assertSucceeds(request)

		By("rebooting the VM")
		request = fmt.Sprintf(`{
		  "method": "reboot_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)

		By("deleting the VM")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})

	It("can create a VM by flavor", func() {
		var vmCID string
		By("creating a VM by flavor")
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "flavor_key_name": "B1_1X2X25",
		      "ephemeral_disk_size": 25,
		      "max_network_speed": 100,
		      "hourly_billing_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "default": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)

		By("deleting the VM")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})

	It("can create a VM with static ip", func() {
		By("creating a VM with static ip")
		var vmCID string
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "default": {
				"cloud_properties": {
				  "vlan_ids": [
					1292651
				  ]
				},
				"dns": [
				  "8.8.8.8"
				],
				"gateway": "10.112.166.129",
				"ip": "10.112.166.30",
				"netmask": "255.255.255.192",
				"type": "manual"
			  },
		      "dynamic": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)

		By("deleting the VM")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})

	It("can create a dedicated VM", func() {
		By("creating a VM with static ip")
		var vmCID string
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": true,
		      "datacenter": "lon02"
		    },
		    {
		      "default": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)

		By("deleting the VM")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})

	It("can create a VM only having private network", func() {
		By("creating a VM having private network")
		var vmCID string
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "dynamic": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)

		By("deleting the VM")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})

	It("creates a VM only having public network and receives an error message", func() {
		By("creating a VM having public network")
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "dynamic": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		resp, err := execCPI(request)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Error.Message).ToNot(BeEmpty())
	})

	It("can create a VM and upgrade the VM", func() {
		By("creating a VM")
		var vmCID string
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "dynamic": {
		        "type": "dynamic",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)
		updateVMID, err := strconv.Atoi(vmCID)

		mask := "id, maxMemory, startCpus, dedicatedAccountHostOnlyFlag, hourlyBillingFlag, datacenter[name], primaryBackendIpAddress, fullyQualifiedDomainName"
		virtualGuest, err := softlayerClient.VirtualGuestService.Id(updateVMID).Mask(mask).GetObject()
		Expect(err).To(BeNil())

		By("can upgrades number of CPU from 1 to 2")
		Expect(*virtualGuest.StartCpus).To(Equal(1))
		request = fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 2,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "dynamic": {
		        "type": "dynamic",
		        "ip": "%s",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId, *virtualGuest.PrimaryBackendIpAddress)
		assertSucceeds(request)

		mask = "id, maxMemory, startCpus, dedicatedAccountHostOnlyFlag, hourlyBillingFlag, datacenter[name], primaryBackendIpAddress, fullyQualifiedDomainName"
		virtualGuest, err = softlayerClient.VirtualGuestService.Id(updateVMID).Mask(mask).GetObject()
		Expect(err).To(BeNil())
		Expect(*virtualGuest.StartCpus).To(Equal(2))

		By("can upgrades memory 1024 to 2048")
		Expect(*virtualGuest.MaxMemory).To(Equal(1024))
		request = fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 2,
		      "memory": 2048,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "dynamic": {
		        "type": "dynamic",
		        "ip": "%s",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId, *virtualGuest.PrimaryBackendIpAddress)
		assertSucceeds(request)

		mask = "id, maxMemory, startCpus, dedicatedAccountHostOnlyFlag, hourlyBillingFlag, datacenter[name], primaryBackendIpAddress, fullyQualifiedDomainName"
		virtualGuest, err = softlayerClient.VirtualGuestService.Id(updateVMID).Mask(mask).GetObject()
		Expect(err).To(BeNil())
		Expect(*virtualGuest.MaxMemory).To(Equal(2048))

		By("can downgrades VM")
		request = fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "bluebosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "cpu": 1,
		      "memory": 1024,
		      "max_network_speed": 100,
		      "ephemeral_disk_size": 20,
		      "hourly_billing_flag": true,
		      "local_disk_flag": true,
		      "dedicated_account_host_only_flag": false,
		      "datacenter": "lon02"
		    },
		    {
		      "dynamic": {
		        "type": "dynamic",
		        "ip": "%s",
		        "dns": [
			      "8.8.8.8"
		        ],
		        "default": [
		          "dns",
		          "gateway"
		        ],
		        "cloud_properties": {
		          "vlan_ids": [1292653, 1292651]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId, *virtualGuest.PrimaryBackendIpAddress)
		assertSucceeds(request)

		mask = "id, maxMemory, startCpus, dedicatedAccountHostOnlyFlag, hourlyBillingFlag, datacenter[name], primaryBackendIpAddress, fullyQualifiedDomainName"
		virtualGuest, err = softlayerClient.VirtualGuestService.Id(updateVMID).Mask(mask).GetObject()
		Expect(err).To(BeNil())
		Expect(*virtualGuest.StartCpus).To(Equal(1))
		Expect(*virtualGuest.MaxMemory).To(Equal(1024))
		Expect(*virtualGuest.HourlyBillingFlag).To(BeTrue())
	})
})
