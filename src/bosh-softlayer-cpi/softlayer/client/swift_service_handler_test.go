package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	boshlogger "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/ncw/swift"
	"github.com/onsi/gomega/ghttp"
	"github.com/softlayer/softlayer-go/session"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("SwiftServiceHandler", func() {
	var (
		err error

		errOutLog   bytes.Buffer
		logger      cpiLog.Logger
		multiLogger api.MultiLogger
		fs          boshsys.FileSystem

		server        *ghttp.Server
		vps           *vpsVm.Client
		slServer      *ghttp.Server
		swiftClient   *swift.Connection
		swiftEndPoint string
		swiftUsername string
		swiftPassword string
		timeoutSec    int
		retries       int
		storageURL    string

		transportHandler *test_helpers.FakeTransportHandler
		sess             *session.Session
		cli              *slClient.ClientManager

		containerName  string
		objectName     string
		objectFilePath string

		imageName string
		note      string
		cluster   string
		osCode    string

		respParas []map[string]interface{}
	)
	BeforeEach(func() {
		// Fake swift server
		server = ghttp.NewServer()
		swiftEndPoint = server.URL()
		swiftEndPoint, err := url.Parse(server.URL())
		Expect(err).To(BeNil())
		// https://lon02.objectstorage.softlayer.net/auth/v1.0/
		_, port, _ := net.SplitHostPort(swiftEndPoint.Host)
		storageURL = "http://localhost:" + port
		swiftUsername = "fake-account:fake-username"
		swiftClient = slClient.NewSwiftClient(storageURL+"/auth/v1.0/", swiftUsername, swiftPassword, timeoutSec, retries)

		//Fake Softlayer server
		slServer = ghttp.NewServer()
		transportHandler = &test_helpers.FakeTransportHandler{
			FakeServer:           slServer,
			SoftlayerAPIEndpoint: slServer.URL(),
			MaxRetries:           retries,
		}

		nanos := time.Now().Nanosecond()
		logger = cpiLog.NewLogger(boshlogger.LevelDebug, strconv.Itoa(nanos))
		multiLogger = api.MultiLogger{Logger: logger, LogBuff: &errOutLog}
		sess = test_helpers.NewFakeSoftlayerSession(transportHandler)
		cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

		containerName = "fake-container"
		objectName = "fake-objectName"
		objectFilePath = "/tmp/fake-objectNameFile"
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("CreateSwiftContainer", func() {
		It("Create container successfully when swfitClient ContainerCreate call successfully", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateSwiftContainer(containerName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Failed to create container when swfitClient ContainerCreate return error", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.CreateSwiftContainer(containerName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Create Swift container"))
		})

		It("Failed to create container when swfitClient is nil", func() {
			swiftClient = nil
			cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

			err := cli.CreateSwiftContainer(containerName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to connect the Swift server due to empty swift client"))
		})
	})

	Describe("DeleteSwiftContainer", func() {
		It("Delete container successfully when swfitClient ContainerCreate call successfully", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteSwiftContainer(containerName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Failed to delete container when swfitClient ContainerCreate return error", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteSwiftContainer(containerName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Delete Swift container"))
		})

		It("Failed to create container when swfitClient is nil", func() {
			swiftClient = nil
			cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

			err := cli.CreateSwiftContainer(containerName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to connect the Swift server due to empty swift client"))
		})
	})

	Describe("UploadSwiftLargeObject", func() {
		BeforeEach(func() {
			fs = boshsys.NewOsFileSystem(boshlogger.New(boshlogger.LevelDebug, log.New(os.Stdout, "", log.LstdFlags), log.New(os.Stderr, "", log.LstdFlags)))
			fs.WriteFileString(objectFilePath, "1")
		})

		It("Failed to upload when swfitClient put LargeObject return error", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.UploadSwiftLargeObject(containerName, objectName, objectFilePath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Initial object uploader"))
		})

		It("Failed to create container when swfitClient is nil", func() {
			swiftClient = nil
			cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

			err := cli.UploadSwiftLargeObject(containerName, objectName, objectFilePath)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to connect the Swift server due to empty swift client"))
		})
	})

	Describe("DeleteSwiftLargeObject", func() {
		It("Delete object successfully when swfitClient LargeObjectDelete call successfully", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteSwiftLargeObject(containerName, objectName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Failed to delete container when swfitClient LargeObjectDelete return error", func() {
			respParas = []map[string]interface{}{
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusOK,
				},
				{
					"auth":        "fake-auth-token",
					"storage_url": storageURL,
					"filename":    "Swift_auth.json",
					"statusCode":  http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			err := cli.DeleteSwiftLargeObject(containerName, objectName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Delete Swift large object"))
		})

		It("Failed to create container when swfitClient is nil", func() {
			swiftClient = nil
			cli = slClient.NewSoftLayerClientManager(sess, vps, swiftClient, logger)

			err := cli.DeleteSwiftLargeObject(containerName, objectName)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to connect the Swift server due to empty swift client"))
		})
	})

	Describe("CreateImageFromExternalSource", func() {
		BeforeEach(func() {
			imageName = "fake-image-name"
			note = "fake-note"
			cluster = "fake-cluster"
			osCode = "fake-oscode"
		})
		Context("when ImageService CreateFromExternalSource call successfully", func() {
			It("create image successfully", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_createObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_setBootMode.json",
						"statusCode": http.StatusOK,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, slServer)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateImageFromExternalSource(imageName, note, cluster, osCode)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when ImageService CreateFromExternalSource call return an error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_createObject_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, slServer)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateImageFromExternalSource(imageName, note, cluster, osCode)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Create image template from external source"))
			})
		})

		Context("when ImageService SetBootMode call return an error", func() {
			It("Return error", func() {
				respParas = []map[string]interface{}{
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_createObject.json",
						"statusCode": http.StatusOK,
					},
					{
						"filename":   "SoftLayer_Virtual_Guest_Block_Device_Template_Group_setBootMode_InternalError.json",
						"statusCode": http.StatusInternalServerError,
					},
				}
				err = test_helpers.SpecifyServerResps(respParas, slServer)
				Expect(err).NotTo(HaveOccurred())

				_, err := cli.CreateImageFromExternalSource(imageName, note, cluster, osCode)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Set boot mode of image template"))
			})
		})
	})
})
