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
	"github.com/softlayer/softlayer-go/sl"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("TicketHandler", func() {
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

		ticketSubject        *string
		ticketTitle          *string
		ticketContent        *string
		ticketAttachmentId   *int
		ticketAttachmentType *string

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

		ticketSubject = sl.String("OS Reload Question")
		ticketTitle = sl.String("fake-ticket-title")
		ticketContent = sl.String("fake-ticket-content")
		ticketAttachmentId = sl.Int(12345678)
		ticketAttachmentType = sl.String("VIRTUAL_GUEST")
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("CreateTicket", func() {
		It("Create ticket successfully", func() {
			respParas = []map[string]interface{}{
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

			err := cli.CreateTicket(ticketSubject, ticketTitle, ticketContent, ticketAttachmentId, ticketAttachmentType)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayer_Ticket_Subject getAllObjects call return error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Ticket_Subject_getAllObjects_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateTicket(ticketSubject, ticketTitle, ticketContent, ticketAttachmentId, ticketAttachmentType)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})

		It("Return error when could not find suitable ticket subject", func() {
			ticketSubject = sl.String("fake-ticket-subject")

			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Ticket_Subject_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateTicket(ticketSubject, ticketTitle, ticketContent, ticketAttachmentId, ticketAttachmentType)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find suitable ticket subject."))
		})

		It("Return error when SoftLayer_Account getCurrentUser call return error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Ticket_Subject_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Account_getCurrentUser_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateTicket(ticketSubject, ticketTitle, ticketContent, ticketAttachmentId, ticketAttachmentType)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Getting current user."))
		})

		It("Return error when SoftLayer_Ticket call createStandardTicket return error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Ticket_Subject_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Account_getCurrentUser.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Ticket_createStandardTicket_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateTicket(ticketSubject, ticketTitle, ticketContent, ticketAttachmentId, ticketAttachmentType)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Creating standard ticket."))
		})

		It("Return error when ticket status is not 'OPEN'", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Ticket_Subject_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Account_getCurrentUser.json",
					"statusCode": http.StatusOK,
				},
				{
					"filename":   "SoftLayer_Ticket_createStandardTicket_NotOpenStatus.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateTicket(ticketSubject, ticketTitle, ticketContent, ticketAttachmentId, ticketAttachmentType)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Ticket status is not 'OPEN'"))
		})
	})
})
