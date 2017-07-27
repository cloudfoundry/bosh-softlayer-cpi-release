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
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
	"time"
)

var _ = Describe("InstanceHandler", func() {
	var (
		server      *ghttp.Server
		vpsEndPoint string
		vps         *vpsVm.Client

		sess *session.Session
		cli  *boslc.ClientManager

		vgID             int
		vlanID           int
		primaryBackendIP string
		primaryIP        string
		allowedHostID    int
		stemcellID       int

		vgTemplate *datatypes.Virtual_Guest
	)
	BeforeEach(func() {
		// the fake server to setup VPS Server
		server = ghttp.NewServer()
		vpsEndPoint = server.URL()
		vps = vpsClient.New(httptransport.New(vpsEndPoint,
			"v2", []string{"http"}), strfmt.Default).VM
		sess = test_helpers.NewFakeSoftlayerSession(server)

		cli = boslc.NewSoftLayerClientManager(sess, vps)
		vgID = 25804753
		vlanID = 1262125
		primaryBackendIP = "10.112.172.240"
		primaryIP = "159.8.144.5"
		allowedHostID = 123456
		stemcellID = 12345678

		vgTemplate = &datatypes.Virtual_Guest{
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

	Describe("GetInstance", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("get instance successfully", func() {
				vgs, succ, err := cli.GetInstance(vgID, boslc.INSTANCE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*vgs.Id).To(Equal(vgID))
			})
		})

		Context("when VirtualGuestService getObject call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetInstance(vgID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	Describe("GetVlan", func() {
		Context("when NetworkVlanService getObject call successfully", func() {
			It("get vlan successfully", func() {
				networkVlan, succ, err := cli.GetVlan(vlanID, boslc.INSTANCE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*networkVlan.Id).To(Equal(vlanID))
			})
		})

		Context("when NetworkVlanService getObject call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetVlan(vlanID, "fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	Describe("GetInstanceByPrimaryBackendIpAddress", func() {
		Context("when AccountService getVirtualGuests call successfully", func() {
			It("get instance by primary backend ip successfully", func() {
				vgs, succ, err := cli.GetInstanceByPrimaryBackendIpAddress(primaryBackendIP)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*vgs.PrimaryBackendIpAddress).To(Equal(primaryBackendIP))
			})
		})

		Context("when AccountService getVirtualGuests call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetInstanceByPrimaryBackendIpAddress("fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	Describe("GetInstanceByPrimaryIpAddress", func() {
		Context("when AccountService getVirtualGuests call successfully", func() {
			It("get instance by primary ip successfully", func() {
				vgs, succ, err := cli.GetInstanceByPrimaryIpAddress(primaryIP)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*vgs.PrimaryIpAddress).To(Equal(primaryIP))
			})
		})

		Context("when AccountService getVirtualGuests call return an error", func() {
			It("return an error", func() {
				_, succ, err := cli.GetInstanceByPrimaryIpAddress("fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(Equal(false))
			})
		})
	})

	Describe("GetAllowedHostCredential", func() {
		Context("when VirtualGuestService getAllowedHost call successfully", func() {
			It("get instance successfully", func() {
				allowedHost, succ, err := cli.GetAllowedHostCredential(allowedHostID)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*allowedHost.Id).To(Equal(allowedHostID))
			})
		})
	})

	Describe("GetAllowedNetworkStorage", func() {
		Context("when VirtualGuestService getAllowedNetworkStorage call successfully", func() {
			It("get network storage allowed virtual guest successfully", func() {
				networkStorages, succ, err := cli.GetAllowedNetworkStorage(vgID)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(len(networkStorages)).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("WaitInstanceUntilReady", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("waiting until instance ready successfully", func() {
				err := cli.WaitInstanceUntilReady(vgID, time.Now())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("WaitInstanceHasNoneActiveTransaction", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("waiting until instance has none active transaction successfully", func() {
				err := cli.WaitInstanceHasNoneActiveTransaction(vgID, time.Now())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CreateInstance", func() {
		Context("when VirtualGuestService createObject call successfully", func() {
			It("create instance successfully", func() {
				vgs, err := cli.CreateInstance(vgTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(*vgs.FullyQualifiedDomainName).To(Equal(*(*vgTemplate).FullyQualifiedDomainName))
			})
		})
	})

	Describe("EditInstance", func() {
		Context("when VirtualGuestService EditObject call successfully", func() {
			It("edit instance successfully", func() {
				_, err := cli.EditInstance(vgID, vgTemplate)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("RebootInstance", func() {
		Context("when VirtualGuestService rebootDefault call successfully", func() {
			It("create instance successfully", func() {
				err := cli.RebootInstance(vgID, false, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService rebootDefault call successfully", func() {
			It("reboot instance successfully", func() {
				err := cli.RebootInstance(vgID, false, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService rebootSoft call successfully", func() {
			It("reboot instance successfully", func() {
				err := cli.RebootInstance(vgID, true, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService rebootHard call successfully", func() {
			It("reboot instance successfully", func() {
				err := cli.RebootInstance(vgID, false, true)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService reboot choice do not exist", func() {
			It("return an error", func() {
				err := cli.RebootInstance(vgID, true, true)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("The reboot type is not existing"))
			})
		})
	})

	Describe("CancelInstance", func() {
		Context("when VirtualGuestService deleteObject call successfully", func() {
			It("cancel instance successfully", func() {
				err := cli.CancelInstance(vgID)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("UpgradeInstance", func() {
		Context("when upgrade instance's cpu", func() {
			It("upgrade instance successfully", func() {
				_, err := cli.UpgradeInstance(vgID, 2, 0, 0, false, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the cpu option does not exist", func() {
				_, err := cli.UpgradeInstance(vgID, 7, 0, 0, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find guest_core option"))
			})

			It("upgrade instance successfully with private", func() {
				_, err := cli.UpgradeInstance(vgID, 2, 0, 0, true, 0)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when upgrade instance's memory", func() {
			It("upgrade instance successfully", func() {
				_, err := cli.UpgradeInstance(vgID, 0, 1024*8, 0, false, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the ram option does not exist", func() {
				_, err := cli.UpgradeInstance(vgID, 0, 133333, 0, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find ram option"))
			})
		})

		Context("when upgrade instance's network speed", func() {
			It("upgrade instance successfully", func() {
				_, err := cli.UpgradeInstance(vgID, 0, 0, 1000, false, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the port_speed option does not exist", func() {
				_, err := cli.UpgradeInstance(vgID, 0, 0, 1431, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find port_speed option"))
			})
		})

		Context("when add instance's additional disk size", func() {
			It("upgrade instance successfully", func() {
				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).NotTo(HaveOccurred())
			})


			It("return an error if the local disk option does not exist", func() {
				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 401)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No proper (LOCAL) disk"))
			})
		})
	})

	Describe("SetTags", func() {
		Context("when VirtualGuestService setTags call successfully", func() {
			It("set tags successfully", func() {
				succeed, err := cli.SetTags(vgID, `"Tag_compiling":  "buildpack_python"`)
				Expect(err).NotTo(HaveOccurred())
				Expect(succeed).To(Equal(true))
			})
		})
	})


	Describe("GetInstanceAllowedHost", func() {
		Context("when VirtualGuestService getAllowedHost call successfully", func() {
			It("get instance successfully", func() {
				allowedHost, succ, err := cli.GetInstanceAllowedHost(allowedHostID)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*allowedHost.Id).To(Equal(allowedHostID))
			})
		})
	})
})
