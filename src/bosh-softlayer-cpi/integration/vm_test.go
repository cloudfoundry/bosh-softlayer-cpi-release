package integration

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VM", func() {

	It("executes the VM lifecycle", func() {
		var vmCID string
		By("creating a VM")
		request := fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "blusbosh-slcpi-integration-test",
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

		By("creating a VM using flavor_key_name")
		request = fmt.Sprintf(`{
		  "method": "create_vm",
		  "arguments": [
		    "45632666-9fb1-422a-af35-2ab6102c5c1b",
		    "%v",
		    {
		      "hostname_prefix": "blusbosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "flavor_key_name": "B1_1X2X25",
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

		By("deleting the VM of flavor")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})
})
