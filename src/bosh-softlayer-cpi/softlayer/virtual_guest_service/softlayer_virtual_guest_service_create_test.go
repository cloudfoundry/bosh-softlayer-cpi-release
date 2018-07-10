package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	cpiLog "bosh-softlayer-cpi/logger"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
)

var _ = Describe("Virtual Guest Service", func() {
	var (
		cli                 *fakeslclient.FakeClient
		uuidGen             *fakeuuid.FakeGenerator
		logger              cpiLog.Logger
		virtualGuestService SoftlayerVirtualGuestService
	)

	BeforeEach(func() {
		cli = &fakeslclient.FakeClient{}
		uuidGen = &fakeuuid.FakeGenerator{}
		logger = cpiLog.NewLogger(boshlog.LevelDebug, "")
		virtualGuestService = NewSoftLayerVirtualGuestService(cli, uuidGen, logger)
	})

	Describe("Call Create", func() {
		var (
			virtualGuest            *datatypes.Virtual_Guest
			enableVps               bool
			stemcellID              int
			stemcellUUID            string
			publicVlanID            int
			privateVlanID           int
			publicNetworkComponent  *datatypes.Virtual_Guest_Network_Component
			privateNetworkComponent *datatypes.Virtual_Guest_Network_Component
			sshKeys                 []int

			createVmID int
		)

		BeforeEach(func() {
			stemcellUUID, _ = uuidGen.Generate()
			publicVlanID = 2234567
			privateVlanID = 2234568
			publicNetworkComponent = &datatypes.Virtual_Guest_Network_Component{
				NetworkVlan: &datatypes.Network_Vlan{
					Id: sl.Int(publicVlanID),
				},
			}

			privateNetworkComponent = &datatypes.Virtual_Guest_Network_Component{
				NetworkVlan: &datatypes.Network_Vlan{
					Id: sl.Int(privateVlanID),
				},
			}

			virtualGuest = &datatypes.Virtual_Guest{
				Hostname:  sl.String("fake-hostname"),
				Domain:    sl.String("fake-domain.com"),
				StartCpus: sl.Int(2),
				MaxMemory: sl.Int(2048),
				Datacenter: &datatypes.Location{
					Name: sl.String("fake-datacenter"),
				},
				BlockDeviceTemplateGroup: &datatypes.Virtual_Guest_Block_Device_Template_Group{
					GlobalIdentifier: sl.String(stemcellUUID),
				},
				PrivateNetworkOnlyFlag:         sl.Bool(false),
				PrimaryNetworkComponent:        publicNetworkComponent,
				PrimaryBackendNetworkComponent: privateNetworkComponent,
			}
			stemcellID = 1234567
			sshKeys = []int{1234568}

			createVmID = 12345678
		})

		Context("when create from vps", func() {
			BeforeEach(func() {
				enableVps = true
			})

			It("returns virtualGuest Id if create instance successful", func() {
				cli.CreateInstanceFromVPSReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(createVmID),
					},
					nil,
				)

				vmID, err := virtualGuestService.Create(virtualGuest, enableVps, stemcellID, sshKeys)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CreateInstanceFromVPSCallCount()).To(Equal(1))
				Expect(cli.CreateInstanceCallCount()).To(Equal(0))
				Expect(vmID).To(Equal(createVmID))
			})

			It("returns error if softLayerClient create instance from VPS call", func() {
				cli.CreateInstanceFromVPSReturns(
					&datatypes.Virtual_Guest{},
					errors.New("fake-client-error"),
				)

				vmID, err := virtualGuestService.Create(virtualGuest, enableVps, stemcellID, sshKeys)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CreateInstanceFromVPSCallCount()).To(Equal(1))
				Expect(cli.CreateInstanceCallCount()).To(Equal(0))
				Expect(vmID).NotTo(Equal(createVmID))
			})
		})

		Context("when create from VirtualGuestService", func() {
			BeforeEach(func() {
				enableVps = false
			})

			It("returns virtualGuest Id if create instance successful", func() {
				cli.CreateInstanceReturns(
					&datatypes.Virtual_Guest{
						Id: sl.Int(createVmID),
					},
					nil,
				)

				vmID, err := virtualGuestService.Create(virtualGuest, enableVps, stemcellID, sshKeys)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.CreateInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CreateInstanceCallCount()).To(Equal(1))
				Expect(vmID).To(Equal(createVmID))
			})

			It("returns error if softLayerClient create instance from VPS call", func() {
				cli.CreateInstanceReturns(
					&datatypes.Virtual_Guest{},
					errors.New("fake-client-error"),
				)

				vmID, err := virtualGuestService.Create(virtualGuest, enableVps, stemcellID, sshKeys)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.CreateInstanceFromVPSCallCount()).To(Equal(0))
				Expect(cli.CreateInstanceCallCount()).To(Equal(1))
				Expect(vmID).NotTo(Equal(createVmID))
			})
		})
	})

	Describe("Call Cleanup", func() {
		var (
			vmID int
		)

		BeforeEach(func() {
			vmID = 12345678
		})

		It("run successfully Id if create instance successful", func() {
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(vmID),
				},
				true,
				nil,
			)
			cli.CancelInstanceReturns(
				nil,
			)

			err := virtualGuestService.CleanUp(vmID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
			Expect(cli.CancelInstanceCallCount()).To(Equal(1))
		})

		It("returns error if softLayerClient create instance from VPS call", func() {
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(vmID),
				},
				true,
				nil,
			)
			cli.CancelInstanceReturns(
				errors.New("fake-client-error"),
			)

			err := virtualGuestService.CleanUp(vmID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
			Expect(cli.CancelInstanceCallCount()).To(Equal(1))
		})

	})
})
