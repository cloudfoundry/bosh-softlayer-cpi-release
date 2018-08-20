package integration

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("dedicated instance", func() {
	It("can create a dedicated VM", func() {
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
		vmCID = assertSucceedsWithResultOrCatchCapacityError(request).(string)
		// If vmCID is empty, skip test cases about dedicated instances
		if vmCID == "" {
			return
		}
		updateVMID, err := strconv.Atoi(vmCID)
		mask := "id, maxMemory, startCpus, dedicatedAccountHostOnlyFlag, hourlyBillingFlag, datacenter[name], primaryBackendIpAddress, fullyQualifiedDomainName"
		virtualGuest, err := softlayerClient.VirtualGuestService.Id(updateVMID).Mask(mask).GetObject()
		Expect(err).To(BeNil())

		By("can update CPU from 1 to 2")
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
		      "dedicated_account_host_only_flag": true,
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

		By("can update memory 1024 to 2048")
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
		      "dedicated_account_host_only_flag": true,
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

		By("can downgrade VM")
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
		      "dedicated_account_host_only_flag": true,
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

		By("deleting the VM")
		request = fmt.Sprintf(`{
		 "method": "delete_vm",
		 "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})
})
