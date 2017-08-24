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
		      "vmNamePrefix": "blusbosh-slcpi-integration-test",
		      "domain": "softlayer.com",
		      "startCpus": 1,
		      "maxMemory": 1024,
		      "maxNetworkSpeed": 100,
		      "hourlyBillingFlag": true,
		      "localDiskFlag": true,
		      "dedicatedAccountHostOnlyFlag": false,
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
		          "networkVlans":
		          [
		            {
                      "vlanId": 524956
		            },{
		              "vlanId": 524954
		            }
		          ]
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
})
