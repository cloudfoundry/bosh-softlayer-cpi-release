package instance_test

import (
	//"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = FDescribe("Virtual Guest Service", func() {
	var (
		cli                 *fakeslclient.FakeClient
		uuidGen             *fakeuuid.FakeGenerator
		logger              boshlog.Logger
		virtualGuestService SoftlayerVirtualGuestService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = boshlog.NewLogger(boshlog.LevelDebug)
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
		})

		It("Configure networks successfully", func() {
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

			_, err := virtualGuestService.ConfigureNetworks(vmID, networks)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
		})

		//It("Return error if softLayerClient SetTags call returns an error", func() {
		//	cli.SetTagsReturns(
		//		false,
		//		errors.New("fake-client-error"),
		//	)
		//
		//	err := virtualGuestService.SetMetadata(vmID, metaData)
		//	Expect(err).To(HaveOccurred())
		//	Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		//	Expect(cli.SetTagsCallCount()).To(Equal(1))
		//})
	})
})
