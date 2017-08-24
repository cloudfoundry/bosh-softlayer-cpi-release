package instance_test

import (
	//"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"fmt"
)

var _ = Describe("Networks", func() {
	var (
		network Network
		tags    Tags
	)

	Describe("Tag Validate", func() {
		It("Validate single tag successfully", func() {
			tags = Tags{"fake-tag"}

			err := tags.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Validate multiple tag successfully", func() {
			tags = Tags{"fake-tag1", "fake-tag2", "fake-tag3"}

			err := tags.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when validate single tag", func() {
			tags = Tags{"fake@tag"}

			err := tags.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Invalid tag '%s': does not comply with RFC1035", tags[0])))
		})

		It("Return nil when validate empty tags", func() {
			tags = Tags{}

			err := tags.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Tag Unique", func() {
		It("Run Unique successfully", func() {
			tags = Tags{"fake-tag1", "fake-tag2", "fake-tag3", "fake-tag3"}

			uniqueTags := tags.Unique()
			Expect(uniqueTags).To(BeEquivalentTo(Tags{"fake-tag1", "fake-tag2", "fake-tag3"}))
		})

	})

	Describe("Dynamic network setting", func() {
		BeforeEach(func() {
			network = Network{
				Type:    "dynamic",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns"},
				Default: []string{"gateway"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: true,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}
		})

		It("Return positive value when call HasDefaultGateway", func() {
			ret := network.HasDefaultGateway()
			Expect(ret).To(BeTrue())
		})

		It("Return positive value when call SourcePolicyRouting", func() {
			ret := network.SourcePolicyRouting()
			Expect(ret).To(BeTrue())
		})

		It("Return positive value when call IsDynamic", func() {
			ret := network.IsDynamic()
			Expect(ret).To(BeTrue())
		})

		It("Return false value when call IsVip", func() {
			ret := network.IsVip()
			Expect(ret).To(BeFalse())
		})

		It("Return false value when call IsManual", func() {
			ret := network.IsManual()
			Expect(ret).To(BeFalse())
		})

		It("Return corret network settings when call AppendDNS", func() {
			expectNetwork := Network{
				Type:    "dynamic",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns", "appended-dns"},
				Default: []string{"gateway"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: true,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}

			net := network.AppendDNS("appended-dns")
			Expect(net).To(Equal(expectNetwork))
		})

		It("Return corret network settings when call AppendDNS empty dns", func() {
			expectNetwork := Network{
				Type:    "dynamic",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns"},
				Default: []string{"gateway"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: true,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}

			net := network.AppendDNS("")
			Expect(net).To(Equal(expectNetwork))
		})
	})

	Describe("Manual network setting", func() {
		BeforeEach(func() {
			network = Network{
				Type:    "manual",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns"},
				Default: []string{"fake-network-default"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: false,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}
		})

		It("Return negative value when call HasDefaultGateway", func() {
			ret := network.HasDefaultGateway()
			Expect(ret).To(BeFalse())
		})

		It("Return negative value when call SourcePolicyRouting", func() {
			ret := network.SourcePolicyRouting()
			Expect(ret).To(BeFalse())
		})

		It("Return negative value when call IsDynamic", func() {
			ret := network.IsDynamic()
			Expect(ret).To(BeFalse())
		})

		It("Return false value when call IsVip", func() {
			ret := network.IsVip()
			Expect(ret).To(BeFalse())
		})

		It("Return true value when call IsManual", func() {
			ret := network.IsManual()
			Expect(ret).To(BeTrue())
		})
	})

	Describe("Validate", func() {
		It("Validate single dynamic network successfully", func() {
			network = Network{
				Type:    "dynamic",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns"},
				Default: []string{"fake-network-default"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: true,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}

			err := network.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Validate single manual network successfully", func() {
			network = Network{
				Type:    "manual",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns"},
				Default: []string{"fake-network-default"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: true,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}

			err := network.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when validate single vip network", func() {
			network = Network{
				Type:    "vip",
				IP:      "10.10.10.10",
				Gateway: "fake-network-gateway",
				Netmask: "fake-network-netmask",
				DNS:     []string{"fake-network-dns"},
				Default: []string{"fake-network-default"},
				CloudProperties: NetworkCloudProperties{
					VlanID:              42345678,
					SourcePolicyRouting: true,
					Tags:                []string{"fake-network-cloud-network-tag"},
				},
			}

			err := network.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Network type 'vip' not supported"))
		})
	})
})
