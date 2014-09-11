package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "bosh/logger"
	fakeuuid "bosh/uuid/fakes"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakebslcpi "github.com/maximilien/bosh-softlayer-cpi/softlayer/cpi/fakes"
	fakestem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"

	wrdn "github.com/cloudfoundry-incubator/garden/warden"
)

var _ = Describe("SoftLayerCreator", func() {
	var (
		uuidGen                *fakeuuid.FakeGenerator
		softLayerClient        *fakebslcpi.FakeClient
		agentEnvServiceFactory *fakevm.FakeAgentEnvServiceFactory
		hostBindMounts         *fakevm.FakeHostBindMounts
		guestBindMounts        *fakevm.FakeGuestBindMounts
		agentOptions           AgentOptions
		logger                 boshlog.Logger
		creator                SoftLayerCreator
	)

	BeforeEach(func() {
		uuidGen = &fakeuuid.FakeGenerator{}
		softLayerClient = fakebslcpi.New()
		agentEnvServiceFactory = &fakevm.FakeAgentEnvServiceFactory{}
		hostBindMounts = &fakevm.FakeHostBindMounts{}
		guestBindMounts = &fakevm.FakeGuestBindMounts{
			EphemeralBindMountPath:  "/fake-guest-ephemeral-bind-mount-path",
			PersistentBindMountsDir: "/fake-guest-persistent-bind-mounts-dir",
		}
		agentOptions = AgentOptions{Mbus: "fake-mbus"}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		creator = NewSoftLayerCreator(
			uuidGen,
			softLayerClient,
			agentEnvServiceFactory,
			hostBindMounts,
			guestBindMounts,
			agentOptions,
			logger,
		)
	})

	Describe("Create", func() {
		var (
			stemcell *fakestem.FakeStemcell
			networks Networks
			env      Environment
		)

		BeforeEach(func() {
			stemcell = fakestem.NewFakeStemcellWithPath(
				"fake-stemcell-id",
				"/fake-stemcell-path",
			)

			networks = Networks{"fake-net-name": Network{}}

			env = Environment{"fake-env-key": "fake-env-value"}
		})

		It("returns created vm", func() {
			uuidGen.GeneratedUuid = "fake-vm-id"

			agentEnvService := &fakevm.FakeAgentEnvService{}
			agentEnvServiceFactory.NewAgentEnvService = agentEnvService

			expectedVM := NewSoftLayerVM(
				"fake-vm-id",
				softLayerClient,
				agentEnvService,
				hostBindMounts,
				guestBindMounts,
				logger,
			)

			vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
			Expect(err).ToNot(HaveOccurred())
			Expect(vm).To(Equal(expectedVM))
		})

		Context("when generating VM id succeeds", func() {
			BeforeEach(func() {
				uuidGen.GeneratedUuid = "fake-vm-id"
			})

			It("returns error if zero networks are provided", func() {
				vm, err := creator.Create("fake-agent-id", stemcell, Networks{}, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected exactly one network; received zero"))
				Expect(vm).To(Equal(SoftLayerVM{}))
			})

			It("returns error if more than one network is provided", func() {
				networks = Networks{"fake-net1": Network{}, "fake-net2": Network{}}

				vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected exactly one network; received multiple"))
				Expect(vm).To(Equal(SoftLayerVM{}))
			})

			It("creates one container with generated VM id", func() {
				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).ToNot(HaveOccurred())

				count := softLayerClient.Connection.CreateCallCount()
				Expect(count).To(Equal(1))

				containerSpec := softLayerClient.Connection.CreateArgsForCall(0)
				Expect(containerSpec.Handle).To(Equal("fake-vm-id"))
			})

			It("creates container with stemcell as its root fs", func() {
				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).ToNot(HaveOccurred())

				containerSpec := softLayerClient.Connection.CreateArgsForCall(0)
				Expect(containerSpec.RootFSPath).To(Equal("/fake-stemcell-path"))
			})

			It("creates container with bind mounted ephemeral disk and persistent root location", func() {
				hostBindMounts.MakeEphemeralPath = "/fake-host-ephemeral-bind-mount-path"
				hostBindMounts.MakePersistentPath = "/fake-host-persistent-bind-mounts-dir"

				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).ToNot(HaveOccurred())

				containerSpec := softLayerClient.Connection.CreateArgsForCall(0)
				Expect(containerSpec.BindMounts).To(Equal(
					[]wrdn.BindMount{
						wrdn.BindMount{
							SrcPath: "/fake-host-ephemeral-bind-mount-path",
							DstPath: "/fake-guest-ephemeral-bind-mount-path",
							Mode:    wrdn.BindMountModeRW,
							Origin:  wrdn.BindMountOriginHost,
						},
						wrdn.BindMount{
							SrcPath: "/fake-host-persistent-bind-mounts-dir",
							DstPath: "/fake-guest-persistent-bind-mounts-dir",
							Mode:    wrdn.BindMountModeRW,
							Origin:  wrdn.BindMountOriginHost,
						},
					},
				))

				Expect(hostBindMounts.MakeEphemeralID).To(Equal("fake-vm-id"))
				Expect(hostBindMounts.MakePersistentID).To(Equal("fake-vm-id"))
			})

			It("returns error if making host ephemeral bind mount fails", func() {
				hostBindMounts.MakeEphemeralErr = errors.New("fake-make-ephemeral-err")

				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-make-ephemeral-err"))
			})

			It("returns error if making host persistent bind mount fails", func() {
				hostBindMounts.MakePersistentErr = errors.New("fake-make-persistent-err")

				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-make-persistent-err"))
			})

			It("creates container with IP address if network is not dynamic", func() {
				networks["fake-net-name"] = Network{
					Type: "not-dynamic",
					IP:   "fake-ip",
				}

				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).ToNot(HaveOccurred())

				containerSpec := softLayerClient.Connection.CreateArgsForCall(0)
				Expect(containerSpec.Network).To(Equal("fake-ip"))
			})

			It("creates container without IP address if network is dynamic", func() {
				networks["fake-net-name"] = Network{
					Type: "dynamic",
					IP:   "fake-ip", // is not usually set
				}

				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).ToNot(HaveOccurred())

				containerSpec := softLayerClient.Connection.CreateArgsForCall(0)
				Expect(containerSpec.Network).To(BeEmpty()) // fake-ip is not used
			})

			It("creates container without properties", func() {
				_, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).ToNot(HaveOccurred())

				containerSpec := softLayerClient.Connection.CreateArgsForCall(0)
				Expect(containerSpec.Properties).To(Equal(wrdn.Properties{}))
			})

			Context("when creating container succeeds", func() {
				var (
					agentEnvService *fakevm.FakeAgentEnvService
				)

				BeforeEach(func() {
					agentEnvService = &fakevm.FakeAgentEnvService{}
					agentEnvServiceFactory.NewAgentEnvService = agentEnvService
					softLayerClient.Connection.CreateReturns("fake-vm-id", nil)
				})

				It("updates container's agent env", func() {
					_, err := creator.Create("fake-agent-id", stemcell, networks, env)
					Expect(err).ToNot(HaveOccurred())

					expectedAgentEnv := NewAgentEnvForVM(
						"fake-agent-id",
						"fake-vm-id",
						networks,
						env,
						agentOptions,
					)

					Expect(agentEnvServiceFactory.NewContainer).ToNot(BeNil()) // todo
					Expect(agentEnvService.UpdateAgentEnv).To(Equal(expectedAgentEnv))
				})

				ItDestroysContainer := func(errMsg string) {
					It("destroys created container", func() {
						_, err := creator.Create("fake-agent-id", stemcell, networks, env)
						Expect(err).To(HaveOccurred())

						count := softLayerClient.Connection.StopCallCount()
						Expect(count).To(Equal(1))

						handle, force := softLayerClient.Connection.StopArgsForCall(0)
						Expect(handle).To(Equal("fake-vm-id"))
						Expect(force).To(BeFalse())
					})

					Context("when destroying created container fails", func() {
						BeforeEach(func() {
							softLayerClient.Connection.StopReturns(errors.New("fake-stop-err"))
						})

						It("returns running error and not destroy error", func() {
							vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring(errMsg))
							Expect(vm).To(Equal(SoftLayerVM{}))
						})
					})
				}

				Context("when container's agent env succeeds", func() {
					It("starts BOSH Agent in the container", func() {
						_, err := creator.Create("fake-agent-id", stemcell, networks, env)
						Expect(err).ToNot(HaveOccurred())

						count := softLayerClient.Connection.RunCallCount()
						Expect(count).To(Equal(1))

						expectedProcessSpec := wrdn.ProcessSpec{
							Path:       "/usr/sbin/runsvdir-start",
							Privileged: true,
						}

						handle, processSpec, processIO := softLayerClient.Connection.RunArgsForCall(0)
						Expect(handle).To(Equal("fake-vm-id"))
						Expect(processSpec).To(Equal(expectedProcessSpec))
						Expect(processIO).To(Equal(wrdn.ProcessIO{}))
					})

					Context("when BOSH Agent fails to start", func() {
						BeforeEach(func() {
							softLayerClient.Connection.RunReturns(nil, errors.New("fake-run-err"))
						})

						It("returns error if starting BOSH Agent fails", func() {
							vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("fake-run-err"))
							Expect(vm).To(Equal(SoftLayerVM{}))
						})

						ItDestroysContainer("fake-run-err")
					})
				})

				Context("when container's agent env update fails", func() {
					BeforeEach(func() {
						agentEnvService.UpdateErr = errors.New("fake-update-err")
					})

					It("returns error because BOSH Agent will fail to start without agent env", func() {
						vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-update-err"))
						Expect(vm).To(Equal(SoftLayerVM{}))
					})

					ItDestroysContainer("fake-update-err")
				})
			})

			Context("when creating container fails", func() {
				BeforeEach(func() {
					softLayerClient.Connection.CreateReturns("fake-vm-id", errors.New("fake-create-err"))
				})

				It("returns error if creating container fails", func() {
					vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-create-err"))
					Expect(vm).To(Equal(SoftLayerVM{}))
				})
			})
		})

		Context("when generating VM id fails", func() {
			BeforeEach(func() {
				uuidGen.GenerateError = errors.New("fake-generate-err")
			})

			It("returns error if generating VM id fails", func() {
				vm, err := creator.Create("fake-agent-id", stemcell, networks, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-generate-err"))
				Expect(vm).To(Equal(SoftLayerVM{}))
			})
		})
	})
})
