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

var _ = Describe("DiskHandler", func() {
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

		diskID            int
		networkConnInfoID int
		orderID           int

		vg        *datatypes.Virtual_Guest
		respParas []map[string]interface{}
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

		diskID = 17336531
		networkConnInfoID = 123456789
		orderID = 11764035

		vg = &datatypes.Virtual_Guest{
			Id:                       sl.Int(12345678),
			Domain:                   sl.String("wilma.org"),
			Hostname:                 sl.String("wilma2"),
			FullyQualifiedDomainName: sl.String("wilma2.wilma.org"),
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
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkStorage, success, err := cli.GetBlockVolumeDetails(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*networkStorage.Id).To(Equal(diskID))
			})

			It("get block volume successfully when pass empty mask string", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkStorage, success, err := cli.GetBlockVolumeDetails(diskID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
				Expect(*networkStorage.Id).To(Equal(diskID))
			})

			It("return empty volume when StorageService getObject return NotFound Error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getObject_NotFound.json",
						"statusCode": http.StatusNotFound,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetBlockVolumeDetails(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})

		Context("when StorageService getObject call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetBlockVolumeDetails(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetBlockVolumeDetailsBySoftLayerAccount", func() {
		Context("when StorageService getIscsiNetworkStorage call successfully", func() {
			It("get iscsi volume instance successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkStorage, err := cli.GetBlockVolumeDetailsBySoftLayerAccount(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(*networkStorage.Id).To(Equal(diskID))
			})

			It("get iscsi volume instance successfully when pass empty mask string", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				networkStorage, err := cli.GetBlockVolumeDetailsBySoftLayerAccount(diskID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(*networkStorage.Id).To(Equal(diskID))
			})

			It("Return an error when volumes is empty", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_Empty.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.GetBlockVolumeDetailsBySoftLayerAccount(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not find volume with id"))
			})

			It("Return an error when more than one volume", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_MoreOne.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.GetBlockVolumeDetailsBySoftLayerAccount(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Exist more than one volume with id"))
			})
		})

		Context("when StorageService getIscsiNetworkStorage call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.GetBlockVolumeDetailsBySoftLayerAccount(diskID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})
		})
	})

	Describe("GetNetworkStorageTarget", func() {
		Context("when StorageService getNetworkConnectionDetails call successfully", func() {
			It("get network storage target successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getNetworkConnectionDetails.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetNetworkStorageTarget(networkConnInfoID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
			})

			It("get block volume successfully when pass empty mask string", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getNetworkConnectionDetails.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetNetworkStorageTarget(networkConnInfoID, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
			})

			It("return empty ipaddress when StorageService getNetworkConnectionDetails return NotFound Error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getNetworkConnectionDetails_NotFound.json",
						"statusCode": http.StatusNotFound,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetNetworkStorageTarget(networkConnInfoID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})

		Context("when StorageService getNetworkConnectionDetails call return an error", func() {
			It("return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_getNetworkConnectionDetails_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, success, err := cli.GetNetworkStorageTarget(networkConnInfoID, slClient.VOLUME_DETAIL_MASK)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(success).To(Equal(false))
			})
		})
	})

	Describe("GetLocationId", func() {
		Context("when LocationService getDatacenters call successfully", func() {
			It("get location id successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.GetLocationId("dal02")
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return an error when datacenter name invalid", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				locationID, err := cli.GetLocationId("datacenter-name-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid datacenter name specified"))
				Expect(locationID).To(Equal(0))
			})
		})

		Context("when LocationService getDatacenters call return an error", func() {
			It("Return an error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				locationID, err := cli.GetLocationId("fake-client-error")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(locationID).To(Equal(0))
			})
		})
	})

	Describe("WaitVolumeProvisioningWithOrderId", func() {
		Context("when AccountService IscsiNetworkStorage call successfully", func() {
			It("wait volume provisioning successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.WaitVolumeProvisioningWithOrderId(orderID, time.Now().Add(1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return timeout error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_Empty.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_Empty.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.WaitVolumeProvisioningWithOrderId(orderID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Waiting volume provisioning with order id of '%d' has time out", orderID)))

			})
		})

		Context("when AccountService IscsiNetworkStorage call return an error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.WaitVolumeProvisioningWithOrderId(orderID, time.Now().Add(1*time.Hour))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting volumes with order id  of"))
			})
		})
	})

	Describe("CancelBlockVolume", func() {
		Context("when BillingService cancelItem call successfully", func() {
			It("cancel block volume successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Billing_Item_cancelItem.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CancelBlockVolume(diskID, "Unit test do cancel volume action", false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when BillingService cancelItem call return an error", func() {
			It("Return error when BillingService return internal error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CancelBlockVolume(diskID, "Unit test do cancel volume action", false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when BillingService return internal error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Account_getIscsiNetworkStorage_MissingKeyField.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CancelBlockVolume(diskID, "Unit test do cancel volume action", false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No billing item is found to cancel"))
			})
		})
	})

	Describe("AuthorizeHostToVolume", func() {
		Context("when StorageService allowAccessFromVirtualGuest call successfully", func() {
			It("Authorize host to volume  successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.AuthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when StorageService allowAccessFromVirtualGuest call return an error", func() {
			It("Return error when occur internal error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.AuthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when occur object not found error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.AuthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find object with id of '%d'", diskID)))
			})

			It("Return error when timeout error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_Blocking.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_GroupAccessControlError.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_False.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_False.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Network_Storage_allowAccessFromVirtualGuest_False.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.AuthorizeHostToVolume(vg, diskID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Authorizing instance with id '%d' to volume with id '%d' time out", *vg.Id, diskID)))
			})
		})
	})

	Describe("DeauthorizeHostToVolume", func() {
		Context("when StorageService removeAccessFromVirtualGuest call successfully", func() {
			It("deauthorize host to volume successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.DeauthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when StorageService removeAccessFromVirtualGuest call return an error", func() {
			It("Return error when occur internal error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.DeauthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})

			It("Return error when occur object not found error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.DeauthorizeHostToVolume(vg, diskID, time.Now().Add(1*time.Hour))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find object with id of '%d'", diskID)))
			})

			It("Return error when timeout error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest_Blocking.json",
						"statusCode": http.StatusInternalServerError,
					},
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest_False.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest_False.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Network_Storage_removeAccessFromVirtualGuest_False.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.DeauthorizeHostToVolume(vg, diskID, time.Now().Add(3000000))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("De-Authorizing instance with id '%d' to volume with id '%d' time out", *vg.Id, diskID)))
			})
		})
	})

	Describe("OrderBlockVolume", func() {
		It("Order successfully when order 250GB volume", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Order successfully when order 250GB volume with 1500 iops", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 1500)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Order successfully when order 500GB volume", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance500.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 500, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Order successfully when order 1000GB volume", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance1000.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 1000, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when order non-Performance_storage_iscsi storage", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("default_storage_iscsi", "dal02", 250, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Block volume storage_type must be either Performance or Endurance"))
		})

		It("Return error when SoftLayerLocationDatacenter call getDatacenter return an error", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid datacenter name specified. Please provide the lower case short name"))
		})

		It("Return error when unable to find price for performance storage", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to find price for performance storage"))
		})

		It("Return error when unable to find disk space price with size", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 1000, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to find disk space price with size"))
		})

		It("Return error when SoftLayerProductPackage call getItemPrices return an error", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})

		It("Return error when itemPrices is empty before filter", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices_Empty.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No proper performance storage (iSCSI volume) for size"))
		})

		It("Return error when itemPrices is empty after filter", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices_HasLocation.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No proper performance storage (iSCSI volume) for size"))
		})

		It("Order successfully when order 250GB volume with 3000 iops", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 3000)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find price for %d iops for the given volume", 3000)))
		})

		It("Order successfully when order 250GB volume with 500 iops", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 500)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find price for %d iops for the given volume", 500)))
		})

		Context("When strconv.Atoi return error", func() {
			It("Return error When strconv.Atoi CapacityRestrictionMaximum return error", func() {
				respParas = []map[string]interface{}{
					// GetLocationId
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
						"statusCode": http.StatusOK,
					},
					// GetPerformanceIscsiPackage
					{
						"filename":   "SoftLayer_Product_Package_getObject_AtoiCapacityRestrictionMaximum.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find price for %d iops for the given volume", 1500)))
			})

			It("Return error When strconv.Atoi CapacityRestrictionMinimum return error", func() {
				respParas = []map[string]interface{}{
					// GetLocationId
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
						"statusCode": http.StatusOK,
					},
					// GetPerformanceIscsiPackage
					{
						"filename":   "SoftLayer_Product_Package_getObject_AtoiCapacityRestrictionMinimum.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.OrderBlockVolume("performance_storage_iscsi", "dal02", 250, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find price for %d iops for the given volume", 1500)))
			})
		})
	})

	Describe("OrderBlockVolume22", func() {
		It("Order successfully when order 250GB volume", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Order successfully when order 250GB volume with 1500 iops", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 1500, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Order successfully when order 500GB volume", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService500.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 500, 0, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Order successfully when order 1000GB volume", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService1000.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 1000, 0, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayerLocationDatacenter call getDatacenter return an error", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid datacenter name specified. Please provide the lower case short name"))
		})

		It("Return error when unable to find price for storage_as_a_service storage", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
		})

		It("Return error when unable to find disk space price with size", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 1000, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
		})

		It("Return error when SoftLayerProductPackage getItemPrices return error", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})

		It("Return error when itemPrices is empty before filter", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices_Empty.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No proper performance storage (iSCSI volume) for size"))
		})

		It("Return error when itemPrices is empty after filter", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices_HasLocation.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No proper performance storage (iSCSI volume) for size"))
		})

		It("Order successfully when order 250GB volume with 3000 iops", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 3000, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find price for storage space size: %d, iops: %d", 250, 3000)))
		})

		It("Order successfully when order 250GB volume with 500 iops", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 500, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Unable to find price for storage space size: %d, iops: %d", 250, 500)))
		})

		Context("When set snapshot", func() {
			It("Order successfully when snapshot size is suitable", func() {
				respParas = []map[string]interface{}{
					// GetLocationId
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
						"statusCode": http.StatusOK,
					},
					// GetStorageAsServicePackage
					{
						"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
						"statusCode": http.StatusOK,
					},
					// selectMaximunIopsItemPriceIdOnSize
					{
						"filename":   "SoftLayer_Product_Package_getItemPrices.json",
						"statusCode": http.StatusOK,
					},
					// PlaceOrder
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 10)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when snapshot size is not suitable", func() {
				respParas = []map[string]interface{}{
					// GetLocationId
					{
						"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
						"statusCode": http.StatusOK,
					},
					// GetStorageAsServicePackage
					{
						"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
						"statusCode": http.StatusOK,
					},
					// selectMaximunIopsItemPriceIdOnSize
					{
						"filename":   "SoftLayer_Product_Package_getItemPrices.json",
						"statusCode": http.StatusOK,
					},
					// PlaceOrder
					{
						"filename":   "SoftLayer_Product_Order_placeOrder.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.OrderBlockVolume2("performance_storage_iscsi", "dal02", 250, 0, 100)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price snapshot space size"))
			})
		})
	})

	Describe("CreateVolume", func() {
		It("Create successfully", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitVolumeProvisioningWithOrderId
				{
					"filename":   "SoftLayer_Account_getIscsiNetworkStorage.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateVolume("dal02", 250, 0, 0)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Create successfully when seet snapshotSpace", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetStorageAsServicePackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder.json",
					"statusCode": http.StatusOK,
				},
				// WaitVolumeProvisioningWithOrderId
				{
					"filename":   "SoftLayer_Account_getIscsiNetworkStorage.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateVolume("dal02", 250, 0, 10)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when call OrderBlockVolume return error", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateVolume("dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid datacenter name specified. Please provide the lower case short name"))
		})

		It("Return error when call placeOrder return receipt without order id", func() {
			respParas = []map[string]interface{}{
				// GetLocationId
				{
					"filename":   "SoftLayer_Location_Datacenter_getDatacenters.json",
					"statusCode": http.StatusOK,
				},
				// GetPerformanceIscsiPackage
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
				// selectMaximunIopsItemPriceIdOnSize
				{
					"filename":   "SoftLayer_Product_Package_getItemPrices.json",
					"statusCode": http.StatusOK,
				},
				// PlaceOrder
				{
					"filename":   "SoftLayer_Product_Order_placeOrder_Without_Orderid.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateVolume("dal02", 250, 0, 0)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No order id returned after placing order with size of"))
		})
	})

	Describe("SetNotes", func() {
		Context("when StorageService editObject call successfully", func() {
			It("set tags successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_editObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				success, err := cli.SetNotes(diskID, `"fake-tag-key": "fake-tag-value"`)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(true))
			})
		})

		Context("when StorageService editObject call return error", func() {
			It("return an softlayer-go unhandled error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_editObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				success, err := cli.SetTags(diskID, `"fake-tag-key": "fake-tag-value"`)
				Expect(err).To(HaveOccurred())
				Expect(success).To(Equal(false))
			})

			It("return an ObjectNotFoundError", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_editObject_NotFound.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				success, err := cli.SetTags(diskID, `"fake-tag-key": "fake-tag-value"`)
				Expect(err).NotTo(HaveOccurred())
				Expect(success).To(Equal(false))
			})
		})
	})
})
