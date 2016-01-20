package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("Networks", func() {
	var (
		networks Networks
	)

	BeforeEach(func() {
		networks = Networks{
			"fake-net1-name": Network{
				Type: "fake-net1-type",

				IP:      "fake-net1-ip",
				Netmask: "fake-net1-netmask",
				Gateway: "fake-net1-gateway",

				DNS:     []string{"fake-net1-dns"},
				Default: []string{"fake-net1-default"},

				CloudProperties: map[string]interface{}{
					"fake-net1-cp-key": "fake-net1-cp-value",
				},
			},
			"fake-net2-name": Network{
				Type: "fake-net2-type",
				IP:   "fake-net2-ip",
			},
			"fake-net3-name": Network{
				Type:          "fake-net3-type",
				IP:            "fake-net3-ip",
				Preconfigured: true,
			},
		}
	})

	Describe("AsVMNetworks", func() {
		It("returns networks for VM", func() {
			expectedVMNetworks := bslcvm.Networks{
				"fake-net1-name": bslcvm.Network{
					Type: "fake-net1-type",

					IP:      "fake-net1-ip",
					Netmask: "fake-net1-netmask",
					Gateway: "fake-net1-gateway",

					DNS:           []string{"fake-net1-dns"},
					Default:       []string{"fake-net1-default"},
					Preconfigured: true,

					CloudProperties: map[string]interface{}{
						"fake-net1-cp-key": "fake-net1-cp-value",
					},
				},
				"fake-net2-name": bslcvm.Network{
					Type:          "fake-net2-type",
					IP:            "fake-net2-ip",
					Preconfigured: true,
				},
				"fake-net3-name": bslcvm.Network{
					Type:          "fake-net3-type",
					IP:            "fake-net3-ip",
					Preconfigured: true,
				},
			}

			Expect(networks.AsVMNetworks()).To(Equal(expectedVMNetworks))
		})
	})
})
