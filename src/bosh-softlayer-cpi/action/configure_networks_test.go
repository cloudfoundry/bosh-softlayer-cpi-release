package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "bosh-softlayer-cpi/action"

	instancefakes "bosh-softlayer-cpi/softlayer/virtual_guest_service/fakes"
	registryfakes "bosh-softlayer-cpi/registry/fakes"
)

var _ = Describe("ConfigureNetworks", func() {
	var (
		err      error
		networks Networks

		vmService      *instancefakes.FakeService
		registryClient *registryfakes.FakeClient

		configureNetworks ConfigureNetworks
	)

	BeforeEach(func() {
		vmService = &instancefakes.FakeService{}
		registryClient = &registryfakes.FakeClient{}
		configureNetworks = NewConfigureNetworks(vmService, registryClient)
	})

	Describe("Run", func() {
		var (
			vmCID VMCID
		)
		BeforeEach(func() {
			vmCID = VMCID(12345678)

			networks = Networks{
				"fake-network-name": Network{
					Type:    "dynamic",
					IP:      "fake-network-ip",
					Gateway: "fake-network-gateway",
					Netmask: "fake-network-netmask",
					DNS:     []string{"fake-network-dns"},
					Default: []string{"fake-network-default"},
					CloudProperties: NetworkCloudProperties{
						VlanID:              32345678,
						SourcePolicyRouting: false,
					},
				},
			}
		})

		It("returns an error because method is deprecated", func() {
			_, err = configureNetworks.Run(vmCID, networks)
			Expect(err).To(HaveOccurred())
		})
	})
})
