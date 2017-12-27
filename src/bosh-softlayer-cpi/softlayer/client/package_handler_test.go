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
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"

	"bosh-softlayer-cpi/api"
	cpiLog "bosh-softlayer-cpi/logger"
	slClient "bosh-softlayer-cpi/softlayer/client"
	vpsVm "bosh-softlayer-cpi/softlayer/vps_service/client/vm"
	"bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("packageHandler", func() {
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

		label       string
		key         string
		fingerPrint string
		sshKeyId    int

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

		label = "fake-label"
		key = "fake-key"
		fingerPrint = "fake-fingerPrint"
		sshKeyId = 12345678
	})

	AfterEach(func() {
		test_helpers.DestroyServer(server)
	})

	Describe("GetPackage", func() {
		It("Get package successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects_performance.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetPackage("performance_storage_iscsi")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayer_Product_Package call getAllObjects return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetPackage("performance_storage_iscsi")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})

		It("Return error when SoftLayer_Product_Package call getAllObjects return an empty object", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects_Empty.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetPackage("performance_storage_iscsi")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No packages were found for"))
		})

		It("Return error when SoftLayer_Product_Package call getAllObjects return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getAllObjects.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetPackage("performance_storage_iscsi")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("More than one packages were found for"))
		})
	})

	Describe("GetPerformanceIscsiPackage", func() {
		It("Get Performance Package successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getObject_Performance.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetPerformanceIscsiPackage()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayer_Product_Package call getObject return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetPerformanceIscsiPackage()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})
	})

	Describe("GetStorageAsServicePackage", func() {
		It("Get StorageAsService Package successfully", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getObject_StorageAsService.json",
					"statusCode": http.StatusOK,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetStorageAsServicePackage()
			Expect(err).NotTo(HaveOccurred())
		})

		It("Return error when SoftLayer_Product_Package call getObject return an error", func() {
			respParas = []map[string]interface{}{
				{
					"filename":   "SoftLayer_Product_Package_getObject_InternalError.json",
					"statusCode": http.StatusInternalServerError,
				},
			}
			err = test_helpers.SpecifyServerResps(respParas, server)
			Expect(err).NotTo(HaveOccurred())

			_, err = cli.GetStorageAsServicePackage()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
		})
	})

	Describe("Category handlers", func() {
		var (
			productPackage *datatypes.Product_Package
		)
		BeforeEach(func() {
			productPackage = &datatypes.Product_Package{
				Id:       sl.Int(222),
				IsActive: sl.Int(1),
				Name:     sl.String("Performance"),
				Items: []datatypes.Product_Item{
					{
						Capacity: sl.Float(250),
						Prices: []datatypes.Product_Item_Price{
							{
								Categories: []datatypes.Product_Item_Category{
									{
										CategoryCode: sl.String("performance_storage_iscsi"),
									},
									{
										CategoryCode: sl.String("performance_storage_space"),
									},
								},
							},
						},
						Attributes: []datatypes.Product_Item_Attribute{
							{
								Id: sl.Int(98764),
							},
						},
					},
					{
						Capacity: sl.Float(1500),
						Prices: []datatypes.Product_Item_Price{
							{
								Categories: []datatypes.Product_Item_Category{
									{
										CategoryCode: sl.String("performance_storage_iscsi"),
									},
									{
										CategoryCode: sl.String("performance_storage_iops"),
									},
								},
								CapacityRestrictionMinimum: sl.String("200"),
								CapacityRestrictionMaximum: sl.String("500"),
								CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
							},
						},
						Attributes: []datatypes.Product_Item_Attribute{
							{
								Id: sl.Int(98766),
							},
						},
						CapacityMaximum: sl.String("500"),
						CapacityMinimum: sl.String("200"),
					},
					{
						KeyName:  sl.String("100_250_GBS"),
						Capacity: sl.Float(250),
						Prices: []datatypes.Product_Item_Price{
							{
								Categories: []datatypes.Product_Item_Category{
									{
										CategoryCode: sl.String("storage_as_a_service"),
									},
									{
										CategoryCode: sl.String("storage_block"),
									},
								},
							},
							{
								Categories: []datatypes.Product_Item_Category{
									{
										CategoryCode: sl.String("performance_storage_space"),
									},
								},
							},
						},
						Attributes: []datatypes.Product_Item_Attribute{
							{
								Id: sl.Int(98786),
							},
						},
						CapacityMinimum: sl.String("100"),
						CapacityMaximum: sl.String("250"),
						ItemCategory: &datatypes.Product_Item_Category{
							CategoryCode: sl.String("performance_storage_space"),
						},
					},
					{
						Capacity: sl.Float(1000),
						Prices: []datatypes.Product_Item_Price{
							{
								Categories: []datatypes.Product_Item_Category{
									{
										CategoryCode: sl.String("performance_storage_iops"),
									},
								},
								CapacityRestrictionMinimum: sl.String("0"),
								CapacityRestrictionMaximum: sl.String("50"),
								CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
							},
						},
						Attributes: []datatypes.Product_Item_Attribute{
							{
								Id: sl.Int(98986),
							},
						},
						ItemCategory: &datatypes.Product_Item_Category{
							CategoryCode: sl.String("performance_storage_iops"),
						},
						CapacityMinimum: sl.String("500"),
						CapacityMaximum: sl.String("2000"),
					},
					{
						Capacity: sl.Float(10),
						Prices: []datatypes.Product_Item_Price{
							{
								Categories: []datatypes.Product_Item_Category{
									{
										CategoryCode: sl.String("storage_snapshot_space"),
									},
								},
								CapacityRestrictionMinimum: sl.String("1000"),
								CapacityRestrictionMaximum: sl.String("2000"),
								CapacityRestrictionType:    sl.String("IOPS"),
							},
						},
					},
				},
			}
		})

		Describe("FindSaaSPriceByCategory", func() {
			It("Find successfully", func() {
				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "performance_storage_iscsi")
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when find unknown_category", func() {
				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "unknown_category")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
			})

			It("Return error when every price.LocationGroupId exists", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iscsi"),
										},
										{
											CategoryCode: sl.String("performance_storage_space"),
										},
									},
									LocationGroupId: sl.Int(123457),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98764),
								},
							},
						},
					},
				}

				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "unknown_category")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
			})
		})

		Describe("FindSaaSPerformSpacePrice", func() {
			It("Find successfully", func() {
				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 250)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when find size by out of scope", func() {
				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
			})

			It("Return error when ItemCategory does not exist", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							KeyName:  sl.String("100_250_GBS"),
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_as_a_service"),
										},
										{
											CategoryCode: sl.String("storage_block"),
										},
									},
								},
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_space"),
										},
									},
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98786),
								},
							},
							CapacityMinimum: sl.String("100"),
							CapacityMaximum: sl.String("250"),
						},
					},
				}

				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 250)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
			})

			It("Return error when CapacityMinimum does not exist", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							KeyName:  sl.String("100_250_GBS"),
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_as_a_service"),
										},
										{
											CategoryCode: sl.String("storage_block"),
										},
									},
								},
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_space"),
										},
									},
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98786),
								},
							},
							CapacityMaximum: sl.String("250"),
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_space"),
							},
						},
					},
				}

				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 250)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
			})

			It("Return error when KeyName mismatched", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							KeyName:  sl.String("GBS"),
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_as_a_service"),
										},
										{
											CategoryCode: sl.String("storage_block"),
										},
									},
								},
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_space"),
										},
									},
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98786),
								},
							},
							CapacityMinimum: sl.String("100"),
							CapacityMaximum: sl.String("250"),
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_space"),
							},
						},
					},
				}

				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 250)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
			})

			It("Return error when LocationGroupId exists", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							KeyName:  sl.String("100_250_GBS"),
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_as_a_service"),
										},
										{
											CategoryCode: sl.String("storage_block"),
										},
									},
									LocationGroupId: sl.Int(22345678),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98786),
								},
							},
							CapacityMinimum: sl.String("100"),
							CapacityMaximum: sl.String("250"),
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_space"),
							},
						},
					},
				}

				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 250)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
			})

			It("Return error when package does not have 'performance_storage_space'", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							KeyName:  sl.String("100_250_GBS"),
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_as_a_service"),
										},
										{
											CategoryCode: sl.String("storage_block"),
										},
									},
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98786),
								},
							},
							CapacityMinimum: sl.String("100"),
							CapacityMaximum: sl.String("250"),
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_space"),
							},
						},
					},
				}

				_, err := slClient.FindSaaSPerformSpacePrice(*productPackage, 250)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage size"))
			})
		})

		Describe("FindSaaSPerformIopsPrice", func() {
			It("Find successfully", func() {
				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 10, 1000)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when volume size out of scope", func() {
				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 100, 1000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for storage space size"))
			})

			It("Return error when iops out of scope", func() {
				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 100, 3000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for storage space size"))
			})

			It("Return error when prices do not have performance_storage_iops category", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(1000),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories:                 []datatypes.Product_Item_Category{},
									CapacityRestrictionMinimum: sl.String("0"),
									CapacityRestrictionMaximum: sl.String("50"),
									CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98986),
								},
							},
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_iops"),
							},
							CapacityMinimum: sl.String("500"),
							CapacityMaximum: sl.String("2000"),
						},
					},
				}

				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 10, 1000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for storage space size"))
			})

			It("Return error when prices do not have LocationGroupId", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(1000),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iops"),
										},
									},
									CapacityRestrictionMinimum: sl.String("0"),
									CapacityRestrictionMaximum: sl.String("50"),
									CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
									LocationGroupId:            sl.Int(32345678),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98986),
								},
							},
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_iops"),
							},
							CapacityMinimum: sl.String("500"),
							CapacityMaximum: sl.String("2000"),
						},
					},
				}

				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 10, 1000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for storage space size"))
			})

			It("Return error when prices do not have LocationGroupId", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(1000),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iops"),
										},
									},
									CapacityRestrictionMinimum: sl.String("0"),
									CapacityRestrictionMaximum: sl.String("50"),
									CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98986),
								},
							},
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_iops"),
							},
							CapacityMaximum: sl.String("2000"),
						},
					},
				}

				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 10, 1000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for storage space size"))
			})

			It("Return error when prices do not have LocationGroupId", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(1000),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iops"),
										},
									},
									CapacityRestrictionMinimum: sl.String("0"),
									CapacityRestrictionMaximum: sl.String("50"),
									CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98986),
								},
							},
							CapacityMinimum: sl.String("500"),
							CapacityMaximum: sl.String("2000"),
						},
					},
				}

				_, err := slClient.FindSaaSPerformIopsPrice(*productPackage, 10, 1000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for storage space size"))
			})
		})

		Describe("FindSaaSPriceByCategory", func() {
			It("Find successfully", func() {
				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "storage_as_a_service")
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when nable to find price storage category", func() {
				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "fake_category")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
			})

			It("Return error when nable to find price storage category", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							KeyName:  sl.String("100_250_GBS"),
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_as_a_service"),
										},
										{
											CategoryCode: sl.String("storage_block"),
										},
									},
									LocationGroupId: sl.Int(42345678),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98786),
								},
							},
							CapacityMinimum: sl.String("100"),
							CapacityMaximum: sl.String("250"),
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_space"),
							},
						},
					},
				}
				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "storage_as_a_service")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
			})
		})

		Describe("FindSaaSSnapshotSpacePrice", func() {
			It("Find successfully", func() {
				_, err := slClient.FindSaaSSnapshotSpacePrice(*productPackage, 10, 1500)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Return error when size out of scope", func() {
				_, err := slClient.FindSaaSSnapshotSpacePrice(*productPackage, 100, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price snapshot space size"))
			})

			It("Return error when iops out of scope", func() {
				_, err := slClient.FindSaaSSnapshotSpacePrice(*productPackage, 10, 500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price snapshot space size"))
			})

			It("Return error when package don't have 'storage_snapshot_space' category", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(10),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories:                 []datatypes.Product_Item_Category{},
									CapacityRestrictionMinimum: sl.String("1000"),
									CapacityRestrictionMaximum: sl.String("2000"),
									CapacityRestrictionType:    sl.String("IOPS"),
								},
							},
						},
					},
				}
				_, err := slClient.FindSaaSSnapshotSpacePrice(*productPackage, 10, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price snapshot space size"))
			})

			It("Return error when package have LocationGroupId item", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(10),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("storage_snapshot_space"),
										},
									},
									CapacityRestrictionMinimum: sl.String("1000"),
									CapacityRestrictionMaximum: sl.String("2000"),
									CapacityRestrictionType:    sl.String("IOPS"),
									LocationGroupId:            sl.Int(52345678),
								},
							},
						},
					},
				}
				_, err := slClient.FindSaaSSnapshotSpacePrice(*productPackage, 10, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price snapshot space size"))
			})
		})

		Describe("FindPerformancePrice", func() {
			It("Find successfully", func() {
				_, err := slClient.FindPerformancePrice(*productPackage, "performance_storage_iscsi")
				Expect(err).NotTo(HaveOccurred())
			})
			It("Return error when find un-existing category", func() {
				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "fake_category")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
			})

			It("Return error when price have LocationGroupId", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iscsi"),
										},
										{
											CategoryCode: sl.String("performance_storage_space"),
										},
									},
									LocationGroupId: sl.Int(52345678),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98764),
								},
							},
						},
					},
				}

				_, err := slClient.FindSaaSPriceByCategory(*productPackage, "performance_storage_iscsi")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price storage category"))
			})
		})

		Describe("FindPerformanceSpacePrice", func() {
			It("Find successfully", func() {
				_, err := slClient.FindPerformanceSpacePrice(*productPackage, 250)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Return error when size is out of scope", func() {
				_, err := slClient.FindPerformanceSpacePrice(*productPackage, 5000)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find disk space price with size"))
			})

			It("Return error when price have LocationGroupId", func() {
				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(250),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iscsi"),
										},
										{
											CategoryCode: sl.String("performance_storage_space"),
										},
									},
									LocationGroupId: sl.Int(52345678),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98764),
								},
							},
						},
					},
				}

				_, err := slClient.FindPerformanceSpacePrice(*productPackage, 250)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find disk space price with size"))
			})
		})

		Describe("FindPerformanceIOPSPrice", func() {
			It("Find successfully", func() {
				_, err := slClient.FindPerformanceIOPSPrice(*productPackage, 250, 1500)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Return error when size is out of scope", func() {
				_, err := slClient.FindPerformanceIOPSPrice(*productPackage, 5000, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for"))
			})

			It("Return error when iops is out of scope", func() {
				_, err := slClient.FindPerformanceIOPSPrice(*productPackage, 250, 100)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for"))
			})

			It("Return error when price have LocationGroupId", func() {

				productPackage = &datatypes.Product_Package{
					Id:       sl.Int(222),
					IsActive: sl.Int(1),
					Name:     sl.String("Performance"),
					Items: []datatypes.Product_Item{
						{
							Capacity: sl.Float(1000),
							Prices: []datatypes.Product_Item_Price{
								{
									Categories: []datatypes.Product_Item_Category{
										{
											CategoryCode: sl.String("performance_storage_iops"),
										},
									},
									CapacityRestrictionMinimum: sl.String("0"),
									CapacityRestrictionMaximum: sl.String("50"),
									CapacityRestrictionType:    sl.String("STORAGE_SPACE"),
									LocationGroupId:            sl.Int(52345678),
								},
							},
							Attributes: []datatypes.Product_Item_Attribute{
								{
									Id: sl.Int(98986),
								},
							},
							ItemCategory: &datatypes.Product_Item_Category{
								CategoryCode: sl.String("performance_storage_iops"),
							},
							CapacityMinimum: sl.String("500"),
							CapacityMaximum: sl.String("2000"),
						},
					},
				}

				_, err := slClient.FindPerformanceIOPSPrice(*productPackage, 250, 1500)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unable to find price for"))
			})
		})
	})

})
