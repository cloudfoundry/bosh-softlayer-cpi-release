package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"bosh-softlayer-cpi/softlayer/client"
	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("Virtual Guest Service", func() {
	var (
		cli                 *fakeslclient.FakeClient
		uuidGen             *fakeuuid.FakeGenerator
		logger              boshlog.Logger
		virtualGuestService SoftlayerVirtualGuestService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		virtualGuestService = NewSoftLayerVirtualGuestService(cli, uuidGen, logger)
	})

	Describe("Call ConfigureNetworks", func() {
		var (
			vmID     int
			networks Networks
		)

		BeforeEach(func() {
			vmID = 12345678
			networks = Networks{
				"fake-network1": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 22345,
					},
					Type: "dynamic",
				},
				"fake-network2": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "dynamic",
				},
			}

			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{
					PrimaryNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
						Id:   sl.Int(22345678),
						Name: sl.String("fake-network1"),
						Port: sl.Int(4344),
						NetworkVlan: &datatypes.Network_Vlan{
							Id: sl.Int(22345),
						},
						PrimaryIpAddress: sl.String("fake-ip-addr1"),
						MacAddress:       sl.String("fake-mac-addr1"),
					},
					PrimaryBackendNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
						Id:   sl.Int(32345678),
						Name: sl.String("fake-network2"),
						Port: sl.Int(4345),
						NetworkVlan: &datatypes.Network_Vlan{
							Id: sl.Int(32345),
						},
						PrimaryIpAddress: sl.String("fake-ip-addr2"),
						MacAddress:       sl.String("fake-mac-addr2"),
					},
				},
				true,
				nil,
			)
		})

		It("Configure networks successfully", func() {
			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient ConfigureNetworks call returns an error", func() {
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{},
				false,
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient ConfigureNetworks call returns non-existing", func() {
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{},
				false,
				nil,
			)

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))

		})

		It("Return error if ubuntu call ComponentByNetworkName return an error", func() {
			networks = Networks{
				"fake-network1": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 223456,
					},
					Type: "dynamic",
				},
				"fake-network2": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "dynamic",
				},
			}

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Mapping network component and name"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})

		It("Return error if using incorrect type field call NormalizeNetworkDefinitions return an error", func() {
			networks = Networks{
				"fake-network1": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "dynamic",
				},
				"fake-network2": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "static",
				},
			}

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Normalizing network definitions"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})

		It("Configure networks successfully if networks settings missing `Type` field", func() {
			networks = Networks{
				"fake-network1": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 22345,
					},
				},
				"fake-network2": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "dynamic",
				},
			}

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})

		It("Return error if ubuntu call NormalizeDynamics return an error", func() {
			networks = Networks{
				"fake-network1": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "dynamic",
				},
				"fake-network2": Network{
					CloudProperties: NetworkCloudProperties{
						VlanID: 32345,
					},
					Type: "dynamic",
				},
			}

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Normalizing dynamic networks definitions"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})
	})

	Describe("Call GetVlan", func() {
		var (
			vlanID int
			mask   string
		)

		BeforeEach(func() {
			vlanID = 32345678
			mask = client.NETWORK_DEFAULT_VLAN_MASK

			cli.GetVlanReturns(
				&datatypes.Network_Vlan{
					Id:              sl.Int(32345678),
					PrimarySubnetId: sl.Int(658644),
					NetworkSpace:    sl.String("PUBLIC"),
				},
				true,
				nil,
			)
		})

		It("Get vlan successfully", func() {
			_, err := virtualGuestService.GetVlan(vlanID, mask)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetVlanCallCount()).To(Equal(1))
		})

		It("Return error if client call GetVlan returns an error", func() {
			cli.GetVlanReturns(
				&datatypes.Network_Vlan{},
				false,
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.GetVlan(vlanID, mask)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetVlanCallCount()).To(Equal(1))
		})

		It("Return error if client call GetVlan returns non-existing", func() {
			cli.GetVlanReturns(
				&datatypes.Network_Vlan{},
				false,
				nil,
			)

			_, err := virtualGuestService.GetVlan(vlanID, mask)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to get vlan details with id"))
			Expect(cli.GetVlanCallCount()).To(Equal(1))
		})
	})

	Describe("Call GetSubnet", func() {
		var (
			vlanID int
			mask   string
		)

		BeforeEach(func() {
			vlanID = 32345678
			mask = client.NETWORK_DEFAULT_SUBNET_MASK

			cli.GetSubnetReturns(
				&datatypes.Network_Subnet{
					Id:            sl.Int(658644),
					NetworkVlanId: sl.Int(32345678),
					AddressSpace:  sl.String("PUBLIC"),
				},
				true,
				nil,
			)
		})

		It("Get vlan successfully", func() {
			_, err := virtualGuestService.GetSubnet(vlanID, mask)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetSubnetCallCount()).To(Equal(1))
		})

		It("Return error if client call GetSubnet returns an error", func() {
			cli.GetSubnetReturns(
				&datatypes.Network_Subnet{},
				false,
				errors.New("fake-client-error"),
			)

			_, err := virtualGuestService.GetSubnet(vlanID, mask)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetSubnetCallCount()).To(Equal(1))
		})

		It("Return error if client call GetSubnet returns non-existing", func() {
			cli.GetSubnetReturns(
				&datatypes.Network_Subnet{},
				false,
				nil,
			)

			_, err := virtualGuestService.GetSubnet(vlanID, mask)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to get subnet details with id"))
			Expect(cli.GetSubnetCallCount()).To(Equal(1))
		})
	})
})
