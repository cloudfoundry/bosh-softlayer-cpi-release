package instance_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"fmt"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	cpiLog "bosh-softlayer-cpi/logger"
	fakeslclient "bosh-softlayer-cpi/softlayer/client/fakes"
	. "bosh-softlayer-cpi/softlayer/virtual_guest_service"
)

var _ = Describe("Virtual Guest Service", func() {
	var (
		err error

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

	Describe("Call AttachEphemeralDisk", func() {
		var (
			vmID     int
			diskSize int
		)

		BeforeEach(func() {
			vmID = 12345678
			diskSize = 1024
		})

		It("Attach successfully", func() {
			cli.AttachSecondDiskToInstanceReturns(
				nil,
			)

			err = virtualGuestService.AttachEphemeralDisk(vmID, diskSize)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.AttachSecondDiskToInstanceCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient attach second disk to instance call returns an error", func() {
			cli.AttachSecondDiskToInstanceReturns(
				errors.New("fake-client-error"),
			)

			err = virtualGuestService.AttachEphemeralDisk(vmID, diskSize)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.AttachSecondDiskToInstanceCallCount()).To(Equal(1))
		})
	})

	Describe("Call AttachDisk", func() {
		var (
			vmID   int
			diskID int

			fakeIpAddrs     string
			fakeStorageName string
			fakeUsername    string
			fakePassword    string
		)

		BeforeEach(func() {
			vmID = 12345678
			diskID = 22345678

			fakeIpAddrs = "fake-ip-address"
			fakeStorageName = "fake-network-storage"
			fakeUsername = "fake-username"
			fakePassword = "fake-password"

			cli.GetBlockVolumeDetailsBySoftLayerAccountReturns(
				datatypes.Network_Storage{
					ServiceResourceBackendIpAddress: sl.String(fakeIpAddrs),
				},
				nil,
			)
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(12345678),
				},
				true,
				nil,
			)
			cli.AuthorizeHostToVolumeReturns(
				true,
				nil,
			)
			cli.GetAllowedHostCredentialReturns(
				&datatypes.Network_Storage_Allowed_Host{
					Name: sl.String(fakeStorageName),
					Credential: &datatypes.Network_Storage_Credential{
						Username: sl.String(fakeUsername),
						Password: sl.String(fakePassword),
					},
				},
				true,
				nil,
			)
		})

		Context("When softlayer client work well", func() {
			It("Attach successfully", func() {
				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(1))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(1))
				Expect(string(targetInfo)).To(BeEquivalentTo(fmt.Sprintf(
					`{"id":"%d","initiator_name":"%s","target":"%s","username":"%s","password":"%s" }`,
					diskID,
					fakeStorageName,
					fakeIpAddrs,
					fakeUsername,
					fakePassword,
				)))
			})
		})

		Context("When softlayer client return error or non-existing", func() {
			It("return error if softlayerClient call GetNetworkStorageTarget return error", func() {
				cli.GetBlockVolumeDetailsBySoftLayerAccountReturns(
					datatypes.Network_Storage{},
					errors.New("fake-client-error"),
				)

				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(0))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(0))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(0))
				Expect(string(targetInfo)).To(BeEquivalentTo(""))
			})

			It("return error if softlayerClient call GetInstance return error", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{},
					false,
					errors.New("fake-client-error"),
				)

				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(0))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(0))
				Expect(string(targetInfo)).To(BeEquivalentTo(""))
			})

			It("return error if softlayerClient call GetInstance return non-existing", func() {
				cli.GetInstanceReturns(
					&datatypes.Virtual_Guest{},
					false,
					nil,
				)

				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(0))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(0))
				Expect(string(targetInfo)).To(BeEquivalentTo(""))
			})

			It("return error if softlayerClient call AuthorizeHostToVolume return error", func() {
				cli.AuthorizeHostToVolumeReturns(
					false,
					errors.New("fake-client-error"),
				)

				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(1))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(0))
				Expect(string(targetInfo)).To(BeEquivalentTo(""))
			})

			It("return error if softlayerClient call GetAllowedHostCredential return error", func() {
				cli.GetAllowedHostCredentialReturns(
					&datatypes.Network_Storage_Allowed_Host{},
					false,
					errors.New("fake-client-error"),
				)

				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-client-error"))
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(1))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(1))
				Expect(string(targetInfo)).To(BeEquivalentTo(""))
			})

			It("return error if softlayerClient call GetAllowedHostCredential return non-existing", func() {
				cli.GetAllowedHostCredentialReturns(
					&datatypes.Network_Storage_Allowed_Host{},
					false,
					nil,
				)

				targetInfo, err := virtualGuestService.AttachDisk(vmID, diskID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("have not allowed access credential"))
				Expect(cli.GetBlockVolumeDetailsBySoftLayerAccountCallCount()).To(Equal(1))
				Expect(cli.GetInstanceCallCount()).To(Equal(1))
				Expect(cli.AuthorizeHostToVolumeCallCount()).To(Equal(1))
				Expect(cli.GetAllowedHostCredentialCallCount()).To(Equal(1))
				Expect(string(targetInfo)).To(BeEquivalentTo(""))
			})
		})

	})

	Describe("Call AttachedDisks", func() {
		var (
			vmID          int
			attachedDisks []string
		)

		BeforeEach(func() {
			vmID = 12345678
		})

		It("Attach successfully", func() {
			cli.GetAllowedNetworkStorageReturns(
				[]string{"22345678", "23345678"},
				true,
				nil,
			)

			attachedDisks, err = virtualGuestService.AttachedDisks(vmID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetAllowedNetworkStorageCallCount()).To(Equal(1))
			Expect(attachedDisks).To(ConsistOf("22345678", "23345678"))
		})

		It("Return error if softLayerClient GetAllowedNetworkStorage call returns an error", func() {
			cli.GetAllowedNetworkStorageReturns(
				[]string{},
				false,
				errors.New("fake-client-error"),
			)

			_, err = virtualGuestService.AttachedDisks(vmID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetAllowedNetworkStorageCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient GetAllowedNetworkStorage call returns non-existing", func() {
			cli.GetAllowedNetworkStorageReturns(
				[]string{},
				false,
				nil,
			)

			_, err = virtualGuestService.AttachedDisks(vmID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(cli.GetAllowedNetworkStorageCallCount()).To(Equal(1))
		})
	})

	Describe("Call DetachDisk", func() {
		var (
			vmID   int
			diskID int
		)

		BeforeEach(func() {
			vmID = 12345678
			diskID = 22345678

			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{
					Id: sl.Int(12345678),
				},
				true,
				nil,
			)
			cli.DeauthorizeHostToVolumeReturns(
				true,
				nil,
			)
		})

		It("Attach successfully", func() {
			err = virtualGuestService.DetachDisk(vmID, diskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
			Expect(cli.DeauthorizeHostToVolumeCallCount()).To(Equal(1))
		})

		It("Return error if softLayerClient GetInstance call returns an error", func() {
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{},
				false,
				errors.New("fake-client-error"),
			)

			err = virtualGuestService.DetachDisk(vmID, diskID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
			Expect(cli.DeauthorizeHostToVolumeCallCount()).To(Equal(0))
		})

		It("Return error if softLayerClient GetInstance call returns non-existing", func() {
			cli.GetInstanceReturns(
				&datatypes.Virtual_Guest{},
				false,
				nil,
			)

			err = virtualGuestService.DetachDisk(vmID, diskID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
			Expect(cli.DeauthorizeHostToVolumeCallCount()).To(Equal(0))
		})

		It("Return error if softLayerClient GetInstance call returns non-existing", func() {
			cli.DeauthorizeHostToVolumeReturns(
				false,
				errors.New("fake-client-error"),
			)

			err = virtualGuestService.DetachDisk(vmID, diskID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-client-error"))
			Expect(cli.GetInstanceCallCount()).To(Equal(1))
			Expect(cli.DeauthorizeHostToVolumeCallCount()).To(Equal(1))
		})
	})
})
