package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boslc "bosh-softlayer-cpi/softlayer/client"
	"bosh-softlayer-cpi/test_helpers"
	"github.com/go-openapi/strfmt"
	"github.com/onsi/gomega/ghttp"

	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/softlayer/softlayer-go/session"
	"time"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
)

var _ = Describe("ImageHandler", func() {
	var (
		server      *ghttp.Server
		vpsEndPoint string
		vps         *vpsVm.Client

		sess *session.Session
		cli  *boslc.ClientManager

		diskID            int
		networkConnInfoID int
		orderID           int

		vg *datatypes.Virtual_Guest
	)
	BeforeEach(func() {
		// the fake server to setup VPS Server
		server = ghttp.NewServer()
		vpsEndPoint = server.URL()
		vps = vpsClient.New(httptransport.New(vpsEndPoint,
			"v2", []string{"http"}), strfmt.Default).VM

		sess = test_helpers.NewFakeSoftlayerSession(server)
		cli = boslc.NewSoftLayerClientManager(sess, vps)

		diskID = 17336531
		networkConnInfoID = 123456789
		orderID = 11764035

		vg = &datatypes.Virtual_Guest{
			Domain:                       sl.String("wilma.org"),
			Hostname:                     sl.String("wilma2"),
			FullyQualifiedDomainName:     sl.String("wilma2.wilma.org"),
			MaxCpu:                       sl.Int(2),
			StartCpus:                    sl.Int(2),
			MaxMemory:                    sl.Int(2048),
			HourlyBillingFlag:            sl.Bool(true),
			OperatingSystemReferenceCode: sl.String("CENTOS_7_64"),
			LocalDiskFlag:                sl.Bool(true),
			DedicatedAccountHostOnlyFlag: sl.Bool(false),
			Datacenter: &datatypes.Location{
				Name: sl.String("par01"),
			},
			NetworkVlans: []datatypes.Network_Vlan{
				datatypes.Network_Vlan{
					Id:           sl.Int(1421725),
					VlanNumber:   sl.Int(1419),
					NetworkSpace: sl.String("PRIVATE"),
				},
				datatypes.Network_Vlan{
					Id:           sl.Int(1421723),
					VlanNumber:   sl.Int(1307),
					NetworkSpace: sl.String("PUBLIC"),
				},
			},
			PrimaryBackendIpAddress: sl.String("10.127.94.175"),
			PrimaryIpAddress:        sl.String("159.8.71.16"),
		}
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("GetBlockVolumeDetails", func() {
		Context("when StorageService getObject call successfully", func() {
			It("get block volume successfully", func() {
				networkStorage, succ, err := cli.GetBlockVolumeDetails(diskID, boslc.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*networkStorage.Id).To(Equal(diskID))
			})
		})

		Context("when StorageService getObject call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetBlockVolumeDetails(diskID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	Describe("GetBlockVolumeDetails2", func() {
		Context("when StorageService getIscsiNetworkStorage call successfully", func() {
			It("get iscsi volume instance successfully", func() {
				networkStorage, succ, err := cli.GetBlockVolumeDetails2(diskID, boslc.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*networkStorage.Id).To(Equal(diskID))
			})
		})

		Context("when StorageService getIscsiNetworkStorage call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetBlockVolumeDetails2(diskID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	Describe("GetNetworkStorageTarget", func() {
		Context("when StorageService getNetworkConnectionDetails call successfully", func() {
			It("get network storage target successfully", func() {
				_, succ, err := cli.GetNetworkStorageTarget(networkConnInfoID, boslc.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
			})
		})

		Context("when StorageService getNetworkConnectionDetails call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetNetworkStorageTarget(networkConnInfoID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	//FDescribe("OrderBlockVolume", func() {
	//	Context("when StorageService getObject call successfully", func() {
	//		It("get instance successfully", func() {
	//			_, succ, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", boslc.VOLUME_DETAIL_MASK)
	//			Expect(err).NotTo(HaveOccurred())
	//			Expect(succ).To(Equal(true))
	//		})
	//	})
	//
	//
	//	Context("when StorageService getObject call return an error", func() {
	//		It("return an error", func() {
	//			_, succ, err := cli.OrderBlockVolume(networkConnInfoID, "fake-client-error")
	//			Expect(err).To(HaveOccurred())
	//			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
	//			Expect(succ).To(Equal(false))
	//		})
	//	})
	//})

	Describe("GetLocationId", func() {
		Context("when LocationService getDatacenters call successfully", func() {
			It("get location id successfully", func() {
				_, err := cli.GetLocationId("dal02")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when LocationService getDatacenters call return an error", func() {
			It("return an error", func() {
				locationID, err := cli.GetLocationId("fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(locationID).To(Equal(0))
			})
		})
	})

	//Describe("GetPackage", func() {
	//	FContext("when PackageService getAllObjects call successfully", func() {
	//		It("get instance successfully", func() {
	//			_, err := cli.GetPackage("performance_storage_iscsi")
	//			Expect(err).NotTo(HaveOccurred())
	//		})
	//	})
	//
	//	Context("when PackageService getAllObjects call return an error", func() {
	//		It("return an error", func() {
	//			locationID, err := cli.GetLocationId("fake-client-error")
	//			Expect(err).To(HaveOccurred())
	//			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
	//			Expect(locationID).To(Equal(0))
	//		})
	//	})
	//})

	//Describe("CreateVolume", func() {
	//	Context("when LocationService getDatacenters call successfully", func() {
	//		It("get instance successfully", func() {
	//			_, err := cli.CreateVolume("dal02", 200, 1500)
	//			Expect(err).NotTo(HaveOccurred())
	//		})
	//	})
	//})

	Describe("WaitVolumeProvisioningWithOrderId", func() {
		Context("when AccountService IscsiNetworkStorage call successfully", func() {
			It("wait volume provisioning successfully", func() {
				_, err := cli.WaitVolumeProvisioningWithOrderId(orderID, time.Now().Add(1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CancelBlockVolume", func() {
		Context("when BillingService cancelItem call successfully", func() {
			It("cancel block volume successfully", func() {
				_, err := cli.CancelBlockVolume(diskID, "Unit test do cancel volume action", false)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("AuthorizeHostToVolume", func() {
		Context("when StorageService allowAccessFromVirtualGuest call successfully", func() {
			It("authorize host to volume  successfully", func() {
				_, err := cli.AuthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("DeauthorizeHostToVolume", func() {
		Context("when StorageService removeAccessFromVirtualGuest call successfully", func() {
			It("deauthorize host to volume successfully", func() {
				_, err := cli.DeauthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
