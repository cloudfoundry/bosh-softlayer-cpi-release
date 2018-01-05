package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"net/http"
	"strconv"
	"time"

	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/ncw/swift"
	"github.com/onsi/gomega/ghttp"
	"github.com/softlayer/softlayer-go/session"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("SnapshotHandler", func() {
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

		diskId     int
		note       string
		snapshotId int

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

		diskId = 12345678
		note = "fake-note"
		snapshotId = 12345678
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("CreateSnapshot", func() {
		Context("when StorageService CreateSnapshot call successfully", func() {
			It("create snapshotId successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_createSnapshot.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateSnapshot(diskId, note)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when StorageService CreateSnapshot call return an error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_createSnapshot_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateSnapshot(diskId, note)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})
		})
	})

	Describe("DeleteSnapshot", func() {
		Context("when StorageService DeleteObject call successfully", func() {
			It("delete snapshotId successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_deleteObject.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err = cli.DeleteSnapshot(snapshotId)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when StorageService CreateSnapshot call return an error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Network_Storage_deleteObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, server)
				Expect(err).NotTo(HaveOccurred())

				err = cli.DeleteSnapshot(snapshotId)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			})
		})
	})

	Describe("EnableSnapshot", func() {
		It("EnableSnapshot successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Network_Storage_enableSnapshots.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err = cli.EnableSnapshot(snapshotId, "HOURLY", 1, 0, 0, "Monday")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayerNetworkStorage enableSnapshots return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Network_Storage_enableSnapshots_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err = cli.EnableSnapshot(snapshotId, "HOURLY", 1, 0, 0, "Monday")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})
	})

	Describe("DisableSnapshots", func() {
		It("DisableSnapshots successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Network_Storage_disableSnapshots.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err = cli.DisableSnapshots(snapshotId, "HOURLY")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayerNetworkStorage disableSnapshots return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Network_Storage_disableSnapshots_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err = cli.DisableSnapshots(snapshotId, "HOURLY")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})
	})

	Describe("RestoreFromSnapshot", func() {
		It("DisableSnapshots successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Network_Storage_restoreFromSnapshot.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err = cli.RestoreFromSnapshot(snapshotId, snapshotId)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayerNetworkStorage disableSnapshots return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Network_Storage_restoreFromSnapshot_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err = cli.RestoreFromSnapshot(snapshotId, snapshotId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})
	})
})
