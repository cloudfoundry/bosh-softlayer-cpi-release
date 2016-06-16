package clients_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	"github.com/cloudfoundry-community/bosh-softlayer-tools/common"

	slclientfakes "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("BMP client", func() {

	var (
		err                           error
		bmpClient                     clients.BmpClient
		fakeHttpClient                *slclientfakes.FakeHttpClient
		fakeServerSpec                clients.ServerSpec
		fakeCloudProperty             []clients.CloudProperty
		fakeCreateBaremetalsInfo      clients.CreateBaremetalsInfo
		fakeProvisioningBaremetalInfo clients.ProvisioningBaremetalInfo
	)

	BeforeEach(func() {
		fakeHttpClient = slclientfakes.NewFakeHttpClient("fake-username", "fake-password")
		Expect(fakeHttpClient).ToNot(BeNil())

		bmpClient = clients.NewBmpClient("fake-username", "fake-password", "http://fake.url.com", fakeHttpClient, "fake-config-path")
		Expect(bmpClient).ToNot(BeNil())
	})

	Describe("#ConfigPath", func() {
		It("returns the ConfigPath", func() {
			configPath := bmpClient.ConfigPath()
			Expect(configPath).To(Equal("fake-config-path"))
		})
	})

	Describe("#Info", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "Info.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns BMP server info", func() {
			info, err := bmpClient.Info()
			Expect(err).ToNot(HaveOccurred())

			Expect(info.Status).To(Equal(200))
			Expect(info.Data).To(Equal(clients.DataInfo{
				Name:    "fake-name",
				Version: "fake-version"}))
		})

		It("fails when BMP server fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.Info()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#Bms", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "Bms.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an array of BaremetalInfo", func() {
			bmsResponse, err := bmpClient.Bms("fake-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(bmsResponse.Status).To(Equal(200))
			Expect(len(bmsResponse.Data)).To(Equal(2))
			Expect(bmsResponse.Data[0]).To(Equal(clients.BaremetalInfo{
				Id:                 0,
				Hostname:           "hostname0",
				Private_ip_address: "private_ip_address0",
				Public_ip_address:  "public_ip_address0",
				Hardware_status:    "hardware_status0",
				Memory:             0,
				Cpu:                0,
				Provision_date:     "2016-01-01T00:00:00-00:00"}))

			Expect(bmsResponse.Data[1]).To(Equal(clients.BaremetalInfo{
				Id:                 1,
				Hostname:           "hostname1",
				Private_ip_address: "private_ip_address1",
				Public_ip_address:  "public_ip_address1",
				Hardware_status:    "hardware_status1",
				Memory:             1,
				Cpu:                1,
				Provision_date:     "2016-01-01T00:00:00-00:00"}))
		})

		It("fails when BMP server /sl/packages fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.Bms("fake-name")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#SlPackages", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "SlPackages.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an array of DataPackage", func() {
			slPackageResponse, err := bmpClient.SlPackages()
			Expect(err).ToNot(HaveOccurred())

			Expect(slPackageResponse.Status).To(Equal(200))
			Expect(len(slPackageResponse.Data.Packages)).To(Equal(2))
			Expect(slPackageResponse.Data.Packages[0]).To(Equal(clients.Package{
				Id:   0,
				Name: "name0"}))
			Expect(slPackageResponse.Data.Packages[1]).To(Equal(clients.Package{
				Id:   1,
				Name: "name1"}))
		})

		It("fails when BMP server /sl/packages fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.SlPackages()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#Stemcells", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "Stemcells.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an array of stemcells", func() {
			stemcellsResponse, err := bmpClient.Stemcells()
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcellsResponse.Status).To(Equal(200))
			Expect(len(stemcellsResponse.Stemcell)).To(Equal(2))
			Expect(stemcellsResponse.Stemcell[0]).To(Equal(
				"fake-stemcell-0"))
			Expect(stemcellsResponse.Stemcell[1]).To(Equal(
				"fake-stemcell-1"))
		})

		It("fails when BMP server /stemcells fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.Stemcells()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#SlPackageOptions", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "SlPackageOptions.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns DataPackageOptions info", func() {
			slPackageOptionsResponse, err := bmpClient.SlPackageOptions("fake-id")
			Expect(err).ToNot(HaveOccurred())

			Expect(slPackageOptionsResponse.Status).To(Equal(200))

			Expect(len(slPackageOptionsResponse.Data.Category)).To(Equal(2))
			Expect(slPackageOptionsResponse.Data.Category[0]).To(Equal(clients.Category{
				Code: "code0",
				Name: "name0",
				Options: []clients.Option{
					clients.Option{Id: 0, Description: "description0"},
					clients.Option{Id: 1, Description: "description1"},
				},
				Required: true}))
			Expect(slPackageOptionsResponse.Data.Category[1]).To(Equal(clients.Category{
				Code: "code1",
				Name: "name1",
				Options: []clients.Option{
					clients.Option{Id: 0, Description: "description0"},
				},
				Required: false}))

			Expect(len(slPackageOptionsResponse.Data.Datacenter)).To(Equal(2))
			Expect(slPackageOptionsResponse.Data.Datacenter[0]).To(Equal(
				"datacenter0 - location0"))
			Expect(slPackageOptionsResponse.Data.Datacenter[1]).To(Equal(
				"datacenter1 - location1"))
		})

		It("fails when BMP server /sl/package/{packageid}/options fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.SlPackageOptions("fake-id")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#Tasks", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "Tasks.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an array of tasks", func() {
			tasksResponse, err := bmpClient.Tasks(10)
			Expect(err).ToNot(HaveOccurred())

			Expect(tasksResponse.Status).To(Equal(200))

			Expect(len(tasksResponse.Data)).To(Equal(2))
			Expect(tasksResponse.Data[0]).To(Equal(clients.Task{
				Id:          0,
				Description: "fake-description-0",
				StartTime:   "fake-start-time-0",
				Status:      "fake-status-0",
				EndTime:     "fake-end-time-0"}))
			Expect(tasksResponse.Data[1]).To(Equal(clients.Task{
				Id:          1,
				Description: "fake-description-1",
				StartTime:   "fake-start-time-1",
				Status:      "fake-status-1",
				EndTime:     "fake-end-time-1"}))
		})

		It("fails when BMP server /tasks?latest= fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.Tasks(10)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#TaskOutput", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "TaskOutput.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an array of task ouput", func() {
			taskOutputResponse, err := bmpClient.TaskOutput(10, "event")
			Expect(err).ToNot(HaveOccurred())

			Expect(taskOutputResponse.Status).To(Equal(200))

			Expect(len(taskOutputResponse.Data)).To(Equal(2))
			Expect(taskOutputResponse.Data[0]).To(Equal("INFO -- event0"))
			Expect(taskOutputResponse.Data[1]).To(Equal("ERROR -- event1"))
		})

		It("fails when BMP server /task/{taskid}/txt/{level} fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.TaskOutput(10, "event")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#taskJsonOutput", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "TaskJsonOutput.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an array of task ouput", func() {
			taskOutputResponse, err := bmpClient.TaskJsonOutput(10, "task")
			Expect(err).ToNot(HaveOccurred())

			Expect(taskOutputResponse.Status).To(Equal(200))

			Expect(taskOutputResponse.Data["info"]).ToNot(BeNil())
		})

		It("fails when BMP server /task/{taskid}/json/{level} fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.TaskOutput(10, "event")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#UpdateState", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "UpdateState.json")

			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an status after updating state", func() {
			updateStateResponse, err := bmpClient.UpdateState("fake-id", "fake-state")
			Expect(err).ToNot(HaveOccurred())

			Expect(updateStateResponse.Status).To(Equal(200))
		})

		It("fails when BMP server /baremetal/{serverId}/{state} fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.UpdateState("fake-id", "fake-state")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#Login", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "Login.json")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an status after logining", func() {
			loginResponse, err := bmpClient.Login("fake-username", "fake-password")
			Expect(err).ToNot(HaveOccurred())

			Expect(loginResponse.Status).To(Equal(200))
		})

		It("fails when BMP server /login/{username}/{password} fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.Login("fake-username", "fake-password")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#CreateBaremetal", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "CreateBaremetals.json")
			Expect(err).ToNot(HaveOccurred())

			fakeServerSpec = clients.ServerSpec{
				Package:       "fake-package",
				Server:        "fake-server",
				Ram:           "fake-ram",
				Disk0:         "fake-disk0",
				PortSpeed:     "fake-portSpeed",
				PublicVlanId:  "fake-publicvlanid",
				PrivateVlanId: "fake-privatevlanid",
				Hourly:        true,
			}

			fakeCloudProperty = []clients.CloudProperty{
				clients.CloudProperty{
					BoshIP:     "fake-boship",
					Datacenter: "fake-datacenter",
					Baremetal:  true,
					ServerSpec: fakeServerSpec,
				}}

			fakeCreateBaremetalsInfo = clients.CreateBaremetalsInfo{
				BaremetalSpecs: fakeCloudProperty,
				Deployment:     "fake-name",
			}
		})

		It("returns an task ID", func() {
			createBaremetalResponse, err := bmpClient.CreateBaremetals(fakeCreateBaremetalsInfo, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(createBaremetalResponse.Status).To(Equal(201))

			Expect(createBaremetalResponse.Data).To(Equal(clients.TaskInfo{
				TaskId: 10}))
		})

		It("fails when BMP create baremetal fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.CreateBaremetals(fakeCreateBaremetalsInfo, false)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#ProvisioningBaremetal", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "CreateBaremetals.json")
			Expect(err).ToNot(HaveOccurred())

			fakeProvisioningBaremetalInfo = clients.ProvisioningBaremetalInfo{
				VmNamePrefix:     "fake-vmprefix",
				Bm_stemcell:      "fake-bm-stemcell",
				Bm_netboot_image: "fake-bm-netboot-image",
			}
		})

		It("returns an task ID", func() {
			createBaremetalsResponse, err := bmpClient.ProvisioningBaremetal(fakeProvisioningBaremetalInfo)
			Expect(err).ToNot(HaveOccurred())

			Expect(createBaremetalsResponse.Status).To(Equal(201))

			Expect(createBaremetalsResponse.Data).To(Equal(clients.TaskInfo{
				TaskId: 10}))
		})

		It("fails when BMP create baremetals fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.ProvisioningBaremetal(fakeProvisioningBaremetalInfo)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#ProvisioningBaremetal", func() {
		BeforeEach(func() {
			fakeHttpClient.DoRawHttpRequestResponse, err = common.ReadJsonTestFixtures("..", "bmp", "CreateBaremetals.json")
			Expect(err).ToNot(HaveOccurred())

			fakeProvisioningBaremetalInfo = clients.ProvisioningBaremetalInfo{
				VmNamePrefix:     "fake-vmprefix",
				Bm_stemcell:      "fake-bm-stemcell",
				Bm_netboot_image: "fake-bm-netboot-image",
			}
		})

		It("returns an task ID", func() {
			createBaremetalResponse, err := bmpClient.ProvisioningBaremetal(fakeProvisioningBaremetalInfo)
			Expect(err).ToNot(HaveOccurred())

			Expect(createBaremetalResponse.Status).To(Equal(201))

			Expect(createBaremetalResponse.Data).To(Equal(clients.TaskInfo{
				TaskId: 10}))
		})

		It("fails when BMP create baremetal fails", func() {
			fakeHttpClient.DoRawHttpRequestError = errors.New("fake-error")

			_, err := bmpClient.ProvisioningBaremetal(fakeProvisioningBaremetalInfo)
			Expect(err).To(HaveOccurred())
		})
	})

})
