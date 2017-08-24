package action_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/action"

	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
)

var _ = Describe("Networks", func() {
	var ()

	BeforeEach(func() {
	})

	Describe("Call AsInstanceServiceNetworks", func() {
		It("", func() {
			networks := Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "10.10.10.10",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanIds:             []int{42345678},
						SourcePolicyRouting: true,
						Tags:                []string{"fake-network-cloud-network-tag"},
					},
				},
			}
			instanceNetworks := networks.AsInstanceServiceNetworks()
			Expect(instanceNetworks)
		})
	})
})
