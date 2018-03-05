package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/ncw/swift"
	"github.com/onsi/gomega/ghttp"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("InstanceHandler", func() {
	var (
		err error

		errOutLog   bytes.Buffer
		logger      cpiLog.Logger
		multiLogger api.MultiLogger

		server      *ghttp.Server
		vps         *vpsVm.Client
		swiftClient *swift.Connection

		transportHandler *test_helpers.FakeTransportHandler
		sess             *session.Session
		cli              *slClient.ClientManager

		vgID             int
		vlanID           int
		subnetID         int
		primaryBackendIP string
		primaryIP        string
		allowedHostID    int
		stemcellID       int
		sshKeyIds        []int

		vgTemplate *datatypes.Virtual_Guest
		respParas  []map[string]interface{}
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
		transportHandler = &test_helpers.FakeTransportHandler{
			FakeServer:           server,
			SoftlayerAPIEndpoint: server.URL(),
			MaxRetries:           3,
		}

		vps = &vpsVm.Client{}
		swiftClient = &swift.Connection{}

		nanos := time.Now().Nanosecond()
		logger = cpiLog.NewLogger(boshlogger.LevelDebug, strconv.Itoa(nanos))
		multiLogger = api.MultiLogger{Logger: logger, LogBuff: &errOutLog}
		sess = test_helpers.NewFakeSoftlayerSession(transportHandler)
		cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

		vgID = 25804753
		vlanID = 1262125
		subnetID = 510674
		primaryBackendIP = "10.112.172.240"
		primaryIP = "159.8.144.5"
		allowedHostID = 123456
		stemcellID = 12345678
		sshKeyIds = []int{2234566}

		vgTemplate = &datatypes.Virtual_Guest{
			Domain:                   sl.String("wilma.org"),
			Hostname:                 sl.String("wilma2"),
			FullyQualifiedDomainName: sl.String("wilma2.wilma.org"),
			MaxCpu:                       sl.Int(2),
			StartCpus:                    sl.Int(2),
			MaxMemory:                    sl.Int(2048),
			HourlyBillingFlag:            sl.Bool(true),
			OperatingSystemReferenceCode: sl.String("UBUNTU_64"),
			LocalDiskFlag:                sl.Bool(true),
			DedicatedAccountHostOnlyFlag: sl.Bool(false),
			Datacenter: &datatypes.Location{
				Name: sl.String("par01"),
			},
			NetworkVlans: []datatypes.Network_Vlan{
				{
					Id:           sl.Int(1421725),
					VlanNumber:   sl.Int(1419),
					NetworkSpace: sl.String("PRIVATE"),
				},
				{
					Id:           sl.Int(1421723),
					VlanNumber:   sl.Int(1307),
					NetworkSpace: sl.String("PUBLIC"),
				},
			},
			PrimaryBackendIpAddress: sl.String("10.127.94.175"),
			PrimaryIpAddress:        sl.String("159.8.71.16"),
			PrimaryBackendNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
				NetworkVlan: &datatypes.Network_Vlan{
					Id: sl.Int(1421725),
				},
			},
			PrimaryNetworkComponent: &datatypes.Virtual_Guest_Network_Component{
				NetworkVlan: &datatypes.Network_Vlan{
					Id: sl.Int(1421723),
				},
			},
		}
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("GetInstance", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("get instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				vgs, success, err := cli.GetInstance(vgID, slClient.INSTANCE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*vgs.Id).To(Equal(vgID))
			})

			It("get instance successfully when pass empty mask", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				vgs, success, err := cli.GetInstance(vgID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*vgs.Id).To(Equal(vgID))
			})
		})

		Context("when VirtualGuestService getObject call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetInstance(vgID, slClient.INSTANCE_DETAIL_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})

			It("return empty object when VirtualGuestService getObject call return an empty object", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				vgs, success, err := cli.GetInstance(vgID, slClient.INSTANCE_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
				Expect((*vgs).Id).To(BeNil())
			})
		})
	})

	Describe("GetVlan", func() {
		Context("when NetworkVlanService getObject call successfully", func() {
			It("get vlan successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Vlan_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkVlan, success, err := cli.GetVlan(vlanID, slClient.NETWORK_DEFAULT_VLAN_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*networkVlan.Id).To(Equal(vlanID))
			})

			It("get vlan successfully when pass empty mask", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Vlan_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkVlan, success, err := cli.GetVlan(vlanID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*networkVlan.Id).To(Equal(vlanID))
			})
		})

		Context("when NetworkVlanService getObject call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Vlan_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetVlan(vlanID, slClient.NETWORK_DEFAULT_VLAN_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})

			It("return an error when return softlayer-go error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Vlan_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetVlan(vlanID, slClient.NETWORK_DEFAULT_VLAN_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetSubnet", func() {
		Context("when NetworkSubnetService getObject call successfully", func() {
			It("get vlan successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Subnet_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkVlan, success, err := cli.GetSubnet(subnetID, slClient.NETWORK_DEFAULT_SUBNET_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*networkVlan.Id).To(Equal(subnetID))
			})

			It("get vlan successfully when pass empty mask", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Subnet_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkVlan, success, err := cli.GetSubnet(subnetID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*networkVlan.Id).To(Equal(subnetID))
			})
		})

		Context("when NetworkSubnetService getObject call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Subnet_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetSubnet(subnetID, "fake-client-error")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})

			It("return an error when return softlayer-go error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Subnet_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetSubnet(subnetID, slClient.NETWORK_DEFAULT_SUBNET_MASK)

				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetInstanceByPrimaryBackendIpAddress", func() {
		Context("when AccountService getVirtualGuests call successfully", func() {
			It("get instance by primary backend ip successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getVirtualGuests.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				vgs, success, err := cli.GetInstanceByPrimaryBackendIpAddress(primaryBackendIP)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*vgs.PrimaryBackendIpAddress).To(Equal(primaryBackendIP))
			})
		})

		Context("when AccountService getVirtualGuests call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getVirtualGuests_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetInstanceByPrimaryBackendIpAddress("fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})
		})

		Context("when AccountService getVirtualGuests call return empty virtual guests", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getVirtualGuests_Empty.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetInstanceByPrimaryBackendIpAddress(primaryBackendIP)

				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetInstanceByPrimaryIpAddress", func() {
		Context("when AccountService getVirtualGuests call successfully", func() {
			It("get instance by primary ip successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getVirtualGuests.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				vgs, success, err := cli.GetInstanceByPrimaryIpAddress(primaryIP)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*vgs.PrimaryIpAddress).To(Equal(primaryIP))
			})
		})

		Context("when AccountService getVirtualGuests call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getVirtualGuests_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetInstanceByPrimaryIpAddress("fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})
		})

		Context("when AccountService getVirtualGuests call return empty slice", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getVirtualGuests_Empty.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetInstanceByPrimaryIpAddress(primaryBackendIP)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetAllowedHostCredential", func() {
		Context("when VirtualGuestService getAllowedHost call successfully", func() {
			It("get allowed host successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedHost.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				allowedHost, success, err := cli.GetAllowedHostCredential(allowedHostID)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*allowedHost.Id).To(Equal(allowedHostID))
			})
		})

		Context("when VirtualGuestService getAllowedHost call return error", func() {
			It("return an error when return softlayer-go error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedHost_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetAllowedHostCredential(allowedHostID)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})

			It("return an error when return unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedHost_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetAllowedHostCredential(allowedHostID)
				Expect(err).To(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetAllowedNetworkStorage", func() {
		Context("when VirtualGuestService getAllowedNetworkStorage call successfully", func() {
			It("get network storage allowed virtual guest successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedNetworkStorage.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkStorages, success, err := cli.GetAllowedNetworkStorage(vgID)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(len(networkStorages)).To(BeNumerically(">=", 1))
			})
		})

		Context("when VirtualGuestService getAllowedNetworkStorage call return error", func() {
			It("return an error when return softlayer-go error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedNetworkStorage_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetAllowedNetworkStorage(vgID)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})

			It("return an error when return unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedNetworkStorage_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetAllowedNetworkStorage(vgID)
				Expect(err).To(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("WaitInstanceUntilReady", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("waiting until instance ready successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReady(vgID, time.Now().Add(1000))
				Expect(err).NotTo(HaveOccurred())
			})

			It("return error when time out", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReady(vgID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Power on virtual guest with id %d Time Out!", vgID)))
			})
		})

		Context("when VirtualGuestService getObject call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReady(vgID, time.Now())
				Expect(err).To(HaveOccurred())
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReady(vgID, time.Now())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("SoftLayer virtual guest '%d' does not exist", vgID)))
			})
		})
	})

	Describe("WaitInstanceUntilReadyWithTicket", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("waiting until instance ready successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReadyWithTicket(vgID, time.Now().Add(1000))
				Expect(err).NotTo(HaveOccurred())
			})

			It("return error and create ticket when time out", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
						"path":       "/SoftLayer_Virtual_Guest/25804753.json",
						"method":     "GET",
					},
					{
						"filename":   "SoftLayer_Ticket_Subject_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Account_getCurrentUser.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Ticket_createStandardTicket.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReadyWithTicket(vgID, time.Now().Add(3*time.Second))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Power on virtual guest with id %d Time Out!", vgID)))
			})
		})

		Context("when VirtualGuestService getObject call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReadyWithTicket(vgID, time.Now())
				Expect(err).To(HaveOccurred())
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceUntilReadyWithTicket(vgID, time.Now())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("SoftLayer virtual guest '%d' does not exist", vgID)))
			})
		})
	})

	Describe("WaitInstanceHasActiveTransaction", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("waiting until instance has active transaction successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasActiveTransaction(vgID, time.Now().Add(1000))
				Expect(err).NotTo(HaveOccurred())
			})

			It("return error when time out", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasActiveTransaction(vgID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting instance with id of '%d' has active transaction time out", vgID))
			})
		})

		Context("when VirtualGuestService getObject call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasActiveTransaction(vgID, time.Now().Add(1000))
				Expect(err).To(HaveOccurred())
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasActiveTransaction(vgID, time.Now().Add(1000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("SoftLayer virtual guest '%d' does not exist", vgID)))
			})
		})
	})

	Describe("WaitInstanceHasNoneActiveTransaction", func() {
		Context("when VirtualGuestService getObject call successfully", func() {
			It("waiting until instance has none active transaction successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasNoneActiveTransaction(vgID, time.Now().Add(1000))
				Expect(err).NotTo(HaveOccurred())
			})

			It("return error when time out", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasNoneActiveTransaction(vgID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting instance with id of '%d' has none active transaction time out", vgID))
			})

			It("return error when instance stay 'RECLAIM_WAIT' transaction", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn_Reclaim.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn_Reclaim.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasNoneActiveTransaction(vgID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("has 'RECLAIM_WAIT' transaction"))
			})
		})

		Context("when VirtualGuestService getObject call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasNoneActiveTransaction(vgID, time.Now().Add(1000))
				Expect(err).To(HaveOccurred())
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.WaitInstanceHasNoneActiveTransaction(vgID, time.Now().Add(1000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("SoftLayer virtual guest '%d' does not exist", vgID)))
			})
		})
	})

	Describe("CreateInstance", func() {
		Context("when VirtualGuestService createObject call successfully", func() {
			It("create instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_createObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				vgs, err := cli.CreateInstance(vgTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(*vgs.FullyQualifiedDomainName).To(Equal(*(*vgTemplate).FullyQualifiedDomainName))
			})
		})

		Context("when VirtualGuestService createObject call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_createObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateInstance(vgTemplate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Creating instance"))
			})
		})

		Context("when VirtualGuestService WaitInstanceUntilReady call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_createObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateInstance(vgTemplate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready"))
			})
		})
	})

	Describe("EditInstance", func() {
		Context("when VirtualGuestService EditObject call successfully", func() {
			It("edit instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.EditInstance(vgID, vgTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(BeTrue())
			})
		})

		Context("when VirtualGuestService EditObject call return error", func() {
			It("edit instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.EditInstance(vgID, vgTemplate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(BeFalse())
			})

			It("edit instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.EditInstance(vgID, vgTemplate)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(BeFalse())
			})
		})

		Context("when VirtualGuestService WaitInstanceUntilReady call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.EditInstance(vgID, vgTemplate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready"))
				Expect(succ).To(BeFalse())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready"))
			})
		})
	})

	Describe("RebootInstance", func() {
		Context("when VirtualGuestService rebootDefault call successfully", func() {
			It("create instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_rebootDefault.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.RebootInstance(vgID, false, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService rebootSoft call successfully", func() {
			It("reboot instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_rebootSoft.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.RebootInstance(vgID, true, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService rebootHard call successfully", func() {
			It("reboot instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_rebootHard.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

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

	Describe("ReloadInstance", func() {
		Context("when VirtualGuestService calls successfully", func() {
			It("reload instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService calls return an error", func() {
			It("Return error when VirtualGuestService firstly getObject call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance has none active transaction before os_reload"))
			})

			It("Return error when VirtualGuestService reloadOperatingSystem call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when VirtualGuestService secondly getObject call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance has active transaction after launching os_reload"))
			})

			It("Return error when VirtualGuestService thirdly getObject call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready after os_reload"))
			})

			It("Return error when VirtualGuestService editObject call return an internal error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Editing VM hostname after OS Reload"))
			})

			It("Return error when VirtualGuestService editObject call return an object not found error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_reloadOperatingSystem.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_editObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.ReloadInstance(vgID, stemcellID, sshKeyIds, "fake-hostname", "fake-domain")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to edit VM hostname after OS Reload"))
			})
		})
	})

	Describe("CancelInstance", func() {
		Context("when VirtualGuestService deleteObject call successfully", func() {
			It("Cancel instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_deleteObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.CancelInstance(vgID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Cancel instance successfully when instance stays 'RECLAIM_WAIT' transaction 1 minutes", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasActiveTxn_Reclaim.json",
						"statusCode": http.StatusOK,
						"path":       "/SoftLayer_Virtual_Guest/25804753.json",
						"method":     "GET",
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.CancelInstance(vgID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when VirtualGuestService getObject call return error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.CancelInstance(vgID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})
		})

		Context("when VirtualGuestService deleteObject call return error", func() {
			It("Return error when VirtualGuestService deleteObject call return internal error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_deleteObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.CancelInstance(vgID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Deleting instance with id '%d'", vgID)))
			})

			It("Return error when VirtualGuestService deleteObject call return falsity", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_deleteObject_False.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err := cli.CancelInstance(vgID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Deleting instance with id '%d' failed", vgID)))
			})
		})
	})

	Describe("UpgradeInstance", func() {
		Context("when upgrade instance's cpu", func() {
			It("upgrade instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 2, 0, 0, false, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the cpu option does not exist", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 7, 0, 0, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find guest_core option"))
			})

			It("upgrade instance successfully with private", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 2, 0, 0, true, 0)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when upgrade instance's memory", func() {
			It("upgrade instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 1024*8, 0, false, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the ram option does not exist", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 133333, 0, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find ram option"))
			})
		})

		Context("when upgrade instance's network speed", func() {
			It("upgrade instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 1000, false, 0)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the port_speed option does not exist", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 1431, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find port_speed option"))
			})
		})

		Context("when add instance's additional disk size", func() {
			It("upgrade instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					// getUpgradeItemPriceForSecondDisk
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).NotTo(HaveOccurred())
			})

			It("upgrade instance successfully when local disk is 'SAN'", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					// getUpgradeItemPriceForSecondDisk
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_SAN.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).NotTo(HaveOccurred())
			})

			It("upgrade instance successfully when presetId is existing", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					// getUpgradeItemPriceForSecondDisk
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_presetId.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the local disk option does not exist", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 401)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No proper (LOCAL) disk"))
			})
		})

		Context("when upgrade none item", func() {
			It("return an error if the local disk option does not exist", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for upgrade"))
			})
		})

		Context("when VirtualGuestService calls return errors", func() {
			It("Return error when SoftLayerProductPackage getAllObjects call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when SoftLayerProductPackage getAllObjects call return an empty object", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects_Empty.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No package found for type"))
			})

			It("Return error when SoftLayerProductPackage getItems call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when SoftLayerVirtualGuest getUpgradeItemPrices call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when SoftLayerVirtualGuest get localDiskFlag call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when SoftLayerVirtualGuest getUpgradeItemPrices call return an empty object", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices_Empty.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 401)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No proper (LOCAL) disk for size"))
			})

			It("Return error when SoftLayerVirtualGuest placeOrder call return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Product_Package_getAllObjects.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Package_getItems.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Product_Order_placeOrder_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.UpgradeInstance(vgID, 0, 0, 0, false, 300)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})
		})
	})

	Describe("SetTags", func() {
		Context("when VirtualGuestService setTags call successfully", func() {
			It("set tags successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setTags.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				success, err := cli.SetTags(vgID, `"Tag_compiling": "buildpack_python"`)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
			})
		})

		Context("when VirtualGuestService setTags call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setTags_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				success, err := cli.SetTags(vgID, `"Tag_compiling": "buildpack_python"`)
				Expect(err).To(HaveOccurred())
				Expect(success).To(Equal(false))
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setTags_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				success, err := cli.SetTags(vgID, `"Tag_compiling": "buildpack_python"`)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetInstanceAllowedHost", func() {
		Context("when VirtualGuestService getAllowedHost call successfully", func() {
			It("get allowed host successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedHost.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				allowedHost, succ, err := cli.GetInstanceAllowedHost(allowedHostID)
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(Equal(true))
				Expect(*allowedHost.Id).To(Equal(allowedHostID))
			})
		})

		Context("when VirtualGuestService getAllowedHost call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_getAllowedHost_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, succ, err := cli.GetInstanceAllowedHost(allowedHostID)
				Expect(err).To(HaveOccurred())
				Expect(succ).To(BeFalse())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setTags_MismatchId.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, succ, err := cli.GetInstanceAllowedHost(allowedHostID)
				Expect(err).To(HaveOccurred())
				Expect(succ).To(BeFalse())
				Expect(err.Error()).To(ContainSubstring("Unable to get allowed host with instance id"))
			})
		})
	})

	Describe("AttachSecondDiskToInstance", func() {
		It("Attach successfully", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitOrderCompleted
				{
					"filename":   "SoftLayer_Billing_Order_getObject.json",
					"statusCode": http.StatusOK,
				},
				// WaitInstanceUntilReady
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.AttachSecondDiskToInstance(vgID, 300)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when call WaitInstanceHasNoneActiveTransaction return an error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.AttachSecondDiskToInstance(vgID, 300)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Waiting until instance has none active transaction before os_reload"))
		})

		It("Return error when call UpgradeInstance return an error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.AttachSecondDiskToInstance(vgID, 300)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Adding second disk with size"))
		})

		It("Return error when call UpgradeInstance return 'a current price was provided for the upgrade order'", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder_Processing.json",
					"statusCode": http.StatusInternalServerError,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getUpgradeRequest.json",
					"statusCode": http.StatusOK,
				},
				// WaitOrderCompleted
				{
					"filename":   "SoftLayer_Billing_Order_getObject.json",
					"statusCode": http.StatusOK,
				},
				// WaitInstanceUntilReady
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.AttachSecondDiskToInstance(vgID, 300)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when call WaitInstanceHasActiveTransaction return an error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitOrderCompleted
				{
					"filename":   "SoftLayer_Billing_Order_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.AttachSecondDiskToInstance(vgID, 300)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Waiting until order placed has been completed after upgrading instance"))
		})

		It("Return error when call WaitInstanceUntilReady return an error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getUpgradeItemPrices.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_localDisk.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitOrderCompleted
				{
					"filename":   "SoftLayer_Billing_Order_getObject.json",
					"statusCode": http.StatusOK,
				},
				// WaitInstanceUntilReady
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.AttachSecondDiskToInstance(vgID, 300)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready after os_reload"))
		})
	})

	Describe("UpgradeInstanceConfig", func() {
		It("Upgrade successfully", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitOrderCompleted
				{
					"filename":   "SoftLayer_Billing_Order_getObject.json",
					"statusCode": http.StatusOK,
				},
				// WaitInstanceUntilReady
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.UpgradeInstanceConfig(vgID, 2, 0, 0, false)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when call WaitInstanceHasNoneActiveTransaction return error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.UpgradeInstanceConfig(vgID, 2, 0, 0, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Waiting until instance has none active transaction before os_reload"))
		})

		It("Return error when call UpgradeInstance return error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.UpgradeInstanceConfig(vgID, 2, 0, 0, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Upgrading configuration to virutal guest of"))
		})

		It("Return error when call WaitOrderCompleted return error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Billing_Order_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.UpgradeInstanceConfig(vgID, 2, 0, 0, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Waiting until order placed has been completed after upgrading instance"))
		})

		It("Return error when call WaitInstanceUntilReady return error", func() {
			respParas = []map[string]interface{}{
				//WaitInstanceHasNoneActiveTransaction
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
					"statusCode": http.StatusOK,
				},
				// UpgradeInstance
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Package_getItems.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitOrderCompleted
				{
					"filename":   "SoftLayer_Billing_Order_getObject.json",
					"statusCode": http.StatusOK,
				},
				// WaitInstanceUntilReady
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.UpgradeInstanceConfig(vgID, 2, 0, 0, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready after os_reload"))
		})
	})

	Describe("SetInstanceMetadata", func() {
		Context("when VirtualGuestService EditObject call successfully", func() {
			It("set instance's metadata successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setUserMetadata.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_HasNoneActiveTxn.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.SetInstanceMetadata(vgID, sl.String("unit-test"))
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(BeTrue())
			})
		})

		Context("when VirtualGuestService EditObject call return error", func() {
			It("edit instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setUserMetadata_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.SetInstanceMetadata(vgID, sl.String("unit-test"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(succ).To(BeFalse())
			})

			It("edit instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setUserMetadata_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.SetInstanceMetadata(vgID, sl.String("unit-test"))
				Expect(err).NotTo(HaveOccurred())
				Expect(succ).To(BeFalse())
			})
		})

		Context("when VirtualGuestService WaitInstanceUntilReady call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_setUserMetadata.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				succ, err := cli.SetInstanceMetadata(vgID, sl.String("unit-test"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready"))
				Expect(succ).To(BeFalse())
				Expect(err.Error()).To(ContainSubstring("Waiting until instance is ready"))
			})
		})
	})
})
