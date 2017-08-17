package integration

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

var _ = Describe("Disk", func() {

	It("executes the disk lifecycle", func() {
		By("creating a VM")
		var vmCID string
		var diskCID string
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
		      "hourlyBillingFlag": true,
		      "localDiskFlag": true,
		      "dedicatedAccountHostOnlyFlag": false,
		      "datacenter": "lon02",
		      "primaryNetworkComponent": {
		        "networkVlan": {
			      "id": 524956
		        }
		      },
		      "primaryBackendNetworkComponent": {
		        "networkVlan": {
		          "id": 524954
		        }
		      }
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
		          "vlanIds": [ 524956, 524954 ]
		        }
		      }
		    },
		    null,
		    {}
		  ]
		}`, existingStemcellId)
		vmCID = assertSucceedsWithResult(request).(string)

		By("creating a disk in the same zone as a VM")
		request = fmt.Sprintf(`{
		  "method": "create_disk",
		  "arguments": [20480, {}, "%v"]
		}`, vmCID)
		diskCID = assertSucceedsWithResult(request).(string)

		//By("attaching the disk")
		//request = fmt.Sprintf(`{
		//  "method": "attach_disk",
		//  "arguments": ["%v", "%v"]
		//}`, vmCID, diskCID)
		//assertSucceeds(request)
		//
		//By("confirming the attachment of a disk")
		//request = fmt.Sprintf(`{
		//  "method": "get_disks",
		//  "arguments": ["%v"]
		//}`, vmCID)
		//disks := toStringArray(assertSucceedsWithResult(request).([]interface{}))
		//Expect(disks).To(ContainElement(diskCID))
		//
		//By("detaching and deleting a disk")
		//request = fmt.Sprintf(`{
		//  "method": "detach_disk",
		//  "arguments": ["%v", "%v"]
		//}`, vmCID, diskCID)
		//assertSucceeds(request)

		By("deleting a disk")
		request = fmt.Sprintf(`{
		  "method": "delete_disk",
		  "arguments": ["%v"]
		}`, diskCID)
		assertSucceeds(request)

		By("deleting the VM")
		request = fmt.Sprintf(`{
		  "method": "delete_vm",
		  "arguments": ["%v"]
		}`, vmCID)
		assertSucceeds(request)
	})
})
