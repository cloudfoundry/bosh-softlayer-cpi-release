package vm_test

import (
	"errors"

	boshlog "bosh/logger"
	fakewrdnclient "github.com/cloudfoundry-incubator/garden/client/fake_warden_client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakedisk "github.com/maximilien/bosh-softlayer-cpi/disk/fakes"
	. "github.com/maximilien/bosh-softlayer-cpi/vm"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/vm/fakes"
)

var _ = Describe("WardenVM", func() {
	var (
		wardenClient    *fakewrdnclient.FakeClient
		agentEnvService *fakevm.FakeAgentEnvService
		hostBindMounts  *fakevm.FakeHostBindMounts
		guestBindMounts *fakevm.FakeGuestBindMounts
		logger          boshlog.Logger
		vm              WardenVM
	)

	BeforeEach(func() {
		wardenClient = fakewrdnclient.New()
		agentEnvService = &fakevm.FakeAgentEnvService{}
		hostBindMounts = &fakevm.FakeHostBindMounts{}
		guestBindMounts = &fakevm.FakeGuestBindMounts{
			EphemeralBindMountPath:  "/fake-guest-ephemeral-bind-mount-path",
			PersistentBindMountsDir: "/fake-guest-persistent-bind-mounts-dir",
		}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		vm = NewWardenVM(
			"fake-vm-id",
			wardenClient,
			agentEnvService,
			hostBindMounts,
			guestBindMounts,
			logger,
		)
	})

	Describe("Delete", func() {
		It("deletes ephemeral bind mount dir", func() {
			err := vm.Delete()
			Expect(err).ToNot(HaveOccurred())

			Expect(hostBindMounts.DeleteEphemeralID).To(Equal("fake-vm-id"))
		})

		Context("when deleting ephemeral bind mount dir succeeds", func() {
			It("deletes persistent bind mounts dir for persistent disks", func() {
				err := vm.Delete()
				Expect(err).ToNot(HaveOccurred())

				Expect(hostBindMounts.DeletePersistentID).To(Equal("fake-vm-id"))
			})

			Context("when deleting persistent bind mounts dir succeeds", func() {
				It("destroys container and returns no error", func() {
					err := vm.Delete()
					Expect(err).ToNot(HaveOccurred())

					Expect(wardenClient.Connection.DestroyCallCount()).To(Equal(1))
					Expect(wardenClient.Connection.DestroyArgsForCall(0)).To(Equal("fake-vm-id"))
				})

				Context("when destroying container fails", func() {
					BeforeEach(func() {
						wardenClient.Connection.DestroyReturns(errors.New("fake-destroy-err"))
					})

					It("returns error", func() {
						err := vm.Delete()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-destroy-err"))
					})
				})
			})

			Context("when deleting persistent bind mounts dir fails", func() {
				BeforeEach(func() {
					hostBindMounts.DeletePersistentErr = errors.New("fake-delete-persistent-err")
				})

				It("returns error", func() {
					err := vm.Delete()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-delete-persistent-err"))
				})

				It("does not delete container", func() {
					err := vm.Delete()
					Expect(err).To(HaveOccurred())

					Expect(wardenClient.Connection.DestroyCallCount()).To(Equal(0))
				})
			})
		})

		Context("when deleting ephemeral bind mount dir fails", func() {
			BeforeEach(func() {
				hostBindMounts.DeleteEphemeralErr = errors.New("fake-delete-ephemeral-err")
			})

			It("returns error", func() {
				err := vm.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-ephemeral-err"))
			})

			It("does not delete container", func() {
				err := vm.Delete()
				Expect(err).To(HaveOccurred())

				Expect(wardenClient.Connection.DestroyCallCount()).To(Equal(0))
			})
		})
	})

	Describe("AttachDisk", func() {
		var (
			disk *fakedisk.FakeDisk
		)

		BeforeEach(func() {
			disk = fakedisk.NewFakeDiskWithPath("fake-disk-id", "/fake-disk-path")
		})

		It("tries to fetch agent env", func() {
			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnvService.FetchCalled).To(BeTrue())
		})

		Context("when fetching agent env succeeds", func() {
			var (
				agentEnv AgentEnv
			)

			BeforeEach(func() {
				agentEnv = AgentEnv{}.AttachPersistentDisk("fake-disk-id2", "/fake-hint-path2")
				agentEnvService.FetchAgentEnv = agentEnv
			})

			It("mounts persistent bind mounts dir", func() {
				err := vm.AttachDisk(disk)
				Expect(err).ToNot(HaveOccurred())

				Expect(hostBindMounts.MountPersistentID).To(Equal("fake-vm-id"))
				Expect(hostBindMounts.MountPersistentDiskID).To(Equal("fake-disk-id"))
				Expect(hostBindMounts.MountPersistentDiskPath).To(Equal("/fake-disk-path"))
			})

			Context("when mounting persistent bind mounts dir succeeds", func() {
				It("updates agent env attaching persistent disk", func() {
					err := vm.AttachDisk(disk)
					Expect(err).ToNot(HaveOccurred())

					// Expected agent env will have additional persistent disk
					expectedAgentEnv := agentEnv.AttachPersistentDisk(
						"fake-disk-id",
						"/fake-guest-persistent-bind-mounts-dir/fake-disk-id",
					)
					Expect(agentEnvService.UpdateAgentEnv).To(Equal(expectedAgentEnv))
				})

				Context("when updating agent env succeeds", func() {
					It("returns without an error", func() {
						err := vm.AttachDisk(disk)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when updating agent env fails", func() {
					It("returns error", func() {
						agentEnvService.UpdateErr = errors.New("fake-update-err")

						err := vm.AttachDisk(disk)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-update-err"))
					})
				})
			})

			Context("when mounting persistent bind mounts dir fails", func() {
				It("returns error", func() {
					hostBindMounts.MountPersistentErr = errors.New("fake-mount-err")

					err := vm.AttachDisk(disk)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-mount-err"))
				})
			})
		})

		Context("when fetching agent env fails", func() {
			It("returns error", func() {
				agentEnvService.FetchErr = errors.New("fake-fetch-err")

				err := vm.AttachDisk(disk)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-fetch-err"))
			})
		})
	})

	Describe("DetachDisk", func() {
		var (
			disk *fakedisk.FakeDisk
		)

		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk("fake-disk-id")
		})

		It("tries to fetch agent env", func() {
			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnvService.FetchCalled).To(BeTrue())
		})

		Context("when fetching agent env succeeds", func() {
			var (
				agentEnv AgentEnv
			)

			BeforeEach(func() {
				agentEnv = AgentEnv{}.AttachPersistentDisk("fake-disk-id", "/fake-hint-path")
				agentEnv = agentEnv.AttachPersistentDisk("fake-disk-id2", "/fake-hint-path2")
				agentEnvService.FetchAgentEnv = agentEnv
			})

			It("unmounts persistent bind mounts dir", func() {
				err := vm.DetachDisk(disk)
				Expect(err).ToNot(HaveOccurred())

				Expect(hostBindMounts.UnmountPersistentID).To(Equal("fake-vm-id"))
				Expect(hostBindMounts.UnmountPersistentDiskID).To(Equal("fake-disk-id"))
			})

			Context("when unmounting persistent bind mounts dir succeeds", func() {
				It("updates agent env detaching persistent disk", func() {
					err := vm.DetachDisk(disk)
					Expect(err).ToNot(HaveOccurred())

					// Expected agent env will not have first persistent disk
					expectedAgentEnv := agentEnv.DetachPersistentDisk("fake-disk-id")
					Expect(agentEnvService.UpdateAgentEnv).To(Equal(expectedAgentEnv))
				})

				Context("when updating agent env succeeds", func() {
					It("returns without an error", func() {
						err := vm.DetachDisk(disk)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when updating agent env fails", func() {
					It("returns error", func() {
						agentEnvService.UpdateErr = errors.New("fake-update-err")

						err := vm.DetachDisk(disk)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-update-err"))
					})
				})
			})

			Context("when unmounting persistent bind mounts dir fails", func() {
				It("returns error", func() {
					hostBindMounts.UnmountPersistentErr = errors.New("fake-unmount-err")

					err := vm.DetachDisk(disk)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-unmount-err"))
				})
			})
		})

		Context("when fetching agent env fails", func() {
			It("returns error", func() {
				agentEnvService.FetchErr = errors.New("fake-fetch-err")

				err := vm.DetachDisk(disk)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-fetch-err"))
			})
		})
	})
})
