package integration

//import (
//	"fmt"
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	"strconv"
//)
//
//var _ = Describe("Stemcell", func() {
//
//	FIt("create stemcell through raw-stemcell file", func() {
//		var stemcellCID string
//		By("creating a stemcell")
//		request := fmt.Sprintf(`{
//		  "method": "create_stemcell",
//		  "arguments": [
//		    "/Users/bjxzi/journey/raw-stemcell/image",
//		    {
//		      "architecture": "x86_64",
//		      "datacenter-name": "lon02",
//		      "infrastructure": "bluemix",
//		      "root_device_name": "/dev/xvda",
//		      "os-code": "UBUNTU_14_64"
//		    }
//		  ]
//		}`)
//		stemcellCID = assertSucceedsWithResult(request).(string)
//		_, err := strconv.Atoi(stemcellCID)
//		Expect(err).NotTo(HaveOccurred())
//	})
//})
