package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/ncw/swift"
	"github.com/onsi/gomega/ghttp"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsClient "bosh-softlayer-cpi/softlayer/vps_service/client"
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
		slServer    *ghttp.Server
		vps         *vpsVm.Client
		vpsEndPoint string
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

		vgTemplate *datatypes.Virtual_Guest

		respParas []map[string]interface{}
	)
	BeforeEach(func() {
		// Fake VPS server
		server = ghttp.NewServer()
		vpsEndPoint = server.URL()
		vpsEndPoint, err := url.Parse(server.URL())
		Expect(err).To(BeNil())
		vps = vpsClient.New(httptransport.New(vpsEndPoint.Host,
			"v2", []string{"http"}), strfmt.Default).VM

		//Fake Softlayer server
		slServer = ghttp.NewServer()
		transportHandler = &test_helpers.FakeTransportHandler{
			FakeServer:           slServer,
			SoftlayerAPIEndpoint: slServer.URL(),
			MaxRetries:           3,
		}

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

	Describe("CreateInstanceFromVPS", func() {
		It("create instance successfully by VirtualGuestService when VirtualGuestService createObject call successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_orderVmByFilter_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
				{
					"filename":   "VPS_addVm.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
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
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			vgs, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).NotTo(HaveOccurred())
			Expect(*vgs.FullyQualifiedDomainName).To(Equal(*(*vgTemplate).FullyQualifiedDomainName))
		})

		It("create instance successfully by vpsService when vpsService OrderVMByFilter call successfully", func() {
			respParas = []map[string]interface{}{
				// OrderVMByFilter
				{
					"filename":   "VPS_orderVmByFilter.json",
					"statusCode": http.StatusOK,
				},
				// UpdateVM
				{
					"filename":   "VPS_updateVm.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				// ReloadInstance
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
				// GetInstance
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			vgs, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).NotTo(HaveOccurred())
			Expect(*vgs.FullyQualifiedDomainName).To(Equal(*(*vgTemplate).FullyQualifiedDomainName))
		})

		It("Return error when vpsService OrderVMByFilter return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_orderVmByFilter_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Ordering vm from pool"))
		})

		It("create instance successfully by VirtualGuestService when VirtualGuestService createObject call return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_orderVmByFilter_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
				{
					"filename":   "VPS_addVm.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Virtual_Guest_createObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Creating VirtualGuest from SoftLayer client"))
		})

		It("Return error when vpsService AddVM return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_orderVmByFilter_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
				{
					"filename":   "VPS_addVm_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
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
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Adding vm into pool"))
		})

		It("Return error when VirtualGuestService ReloadInstance return an error", func() {
			respParas = []map[string]interface{}{
				// OrderVMByFilter
				{
					"filename":   "VPS_orderVmByFilter.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				// ReloadInstance
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reloading vm from pool"))
		})

		It("Return error when VirtualGuestService GetInstance return an error", func() {
			respParas = []map[string]interface{}{
				// OrderVMByFilter
				{
					"filename":   "VPS_orderVmByFilter.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				// ReloadInstance
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
				// GetInstance
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})

		It("Return error when VirtualGuestService ReloadInstance return an empty object", func() {
			respParas = []map[string]interface{}{
				// OrderVMByFilter
				{
					"filename":   "VPS_orderVmByFilter.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				// ReloadInstance
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
				// GetInstance
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not exist"))
		})

		It("Return error when vpsService UpdateVM return an error", func() {
			respParas = []map[string]interface{}{
				// OrderVMByFilter
				{
					"filename":   "VPS_orderVmByFilter.json",
					"statusCode": http.StatusOK,
				},
				// UpdateVM
				{
					"filename":   "VPS_updateVm_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				// ReloadInstance
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
				// GetInstance
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			_, err := cli.CreateInstanceFromVPS(vgTemplate, stemcellID, []int{12345678})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Updating the hostname of vm"))
		})
	})

	Describe("DeleteInstanceFromVPS", func() {
		It("delete instance successfully when VirtualGuestService GetObject call successfully and vpsService GetVMByCid return not found error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_getVmByCid_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
				{
					"filename":   "VPS_addVm.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteInstanceFromVPS(vgID)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when vpsService getVMByCid call return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_getVmByCid_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteInstanceFromVPS(vgID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Removing vm from pool"))
		})

		It("Return error when VirtualGuestService GetObject call return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_getVmByCid_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteInstanceFromVPS(vgID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Getting virtual guest"))
		})

		It("Return error when vpsService addVm call return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_getVmByCid_NotFound.json",
					"statusCode": http.StatusNotFound,
				},
				{
					"filename":   "VPS_addVm_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Virtual_Guest_getObject.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, slServer)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteInstanceFromVPS(vgID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("Adding vm %d to pool", vgID)))
		})

		It("delete instance successfully when VirtualGuestService GetVMByCid call successfully and vpsService UpdateVmWithState call successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "VPS_getVmByCid.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "VPS_updateVmWithState_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteInstanceFromVPS(vgID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Updating state of vm "))
		})
	})
})
