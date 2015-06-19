package vm_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"

	fakedisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk/fakes"
	fakevm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm/fakes"
	fakesutil "github.com/maximilien/bosh-softlayer-cpi/util/fakes"
	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"
)

var _ = Describe("SoftLayerVM", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		sshClient       *fakesutil.FakeSshClient
		agentEnvService *fakevm.FakeAgentEnvService
		logger          boshlog.Logger
		vm              SoftLayerVM
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")

		agentEnvService = &fakevm.FakeAgentEnvService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)

		vm = NewSoftLayerVM(1234, softLayerClient, sshClient, agentEnvService, logger)
	})

	Describe("Delete", func() {
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("true")
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)
			})

			It("deletes the VM successfully", func() {
				err := vm.Delete()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("invalid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("false")
				vm = NewSoftLayerVM(00000, softLayerClient, sshClient, agentEnvService, logger)
			})

			It("fails deleting the VM", func() {
				err := vm.Delete()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Reboot", func() {
		Context("valid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("true")
				vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)
			})

			It("reboots the VM successfully", func() {
				err := vm.Reboot()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("invalid VM ID is used", func() {
			BeforeEach(func() {
				softLayerClient.DoRawHttpRequestResponse = []byte("false")
				vm = NewSoftLayerVM(00000, softLayerClient, sshClient, agentEnvService, logger)
			})

			It("fails rebooting the VM", func() {
				err := vm.Reboot()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	FDescribe("SetMetadata", func() {
		var (
			metadata VMMetadata
		)

		Context("no tags found in metadata", func() {
			BeforeEach(func() {
				metadataBytes := []byte(`{
				  "director": "fake-director-uuid",
				  "name": "fake-director"
				}`)

				metadata = bslcvm.VMMetadata{}
				err := json.Unmarshal(metadataBytes, &metadata)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not set any tag values on the VM", func() {
				err := vm.SetMetadata(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(softLayerClient.DoRawHttpRequestResponseCount).To(Equal(0))
			})
		})

		Context("found tags in metadata", func() {
			BeforeEach(func() {
				metadataBytes := []byte(`{
				  "director": "fake-director-uuid",
				  "name": "fake-director",
				  "tags": "test, tag, director"
				}`)

				metadata = bslcvm.VMMetadata{}
				err := json.Unmarshal(metadataBytes, &metadata)
				Expect(err).ToNot(HaveOccurred())
			})

			It("the tags value is empty", func() {
				metadata["tags"] = ""
				err := vm.SetMetadata(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(softLayerClient.DoRawHttpRequestResponseCount).To(Equal(0))
			})

			It("at least one tag found", func() {
				err := vm.SetMetadata(metadata)

				Expect(err).ToNot(HaveOccurred())
				Expect(softLayerClient.DoRawHttpRequestResponseCount).To(Equal(1))
			})

			Context("when SLVG.SetTags fails", func() {
				BeforeEach(func() {
					softLayerClient.DoRawHttpRequestResponseError = error.New("fake-error")
				})

				It("fails with error", func() {
					err := vm.SetMetadata(metadata)

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("ConfigureNetworks", func() {
		var (
			networks Networks
		)

		BeforeEach(func() {
			networks = Networks{}
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)
		})

		It("returns NotSupportedError", func() {
			err := vm.ConfigureNetworks(networks)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Not supported"))
			Expect(err.(NotSupportedError).Type()).To(Equal("Bosh::Clouds::NotSupported"))
		})
	})

	Describe("#AttachDisk", func() {
		var (
			disk bslcdisk.Disk
		)
		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk(1234)
			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_getObject.json",
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
		})

		It("attaches the iSCSI volume successfully", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)

			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to attach the iSCSI volume", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", errors.New("fake-error"))
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)

			err := vm.AttachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#DetachDisk", func() {
		var (
			disk bslcdisk.Disk
		)
		BeforeEach(func() {
			disk = fakedisk.NewFakeDisk(1234)
			fileNames := []string{
				"SoftLayer_Virtual_Guest_Service_getObject.json",
				"SoftLayer_Network_Storage_Service_getIscsiVolume.json",
			}
			testhelpers.SetTestFixturesForFakeSoftLayerClient(softLayerClient, fileNames)
		})

		It("detaches the iSCSI volume successfully", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", nil)
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)

			err := vm.DetachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
		})

		It("reports error when failed to detach the iSCSI volume", func() {
			sshClient = fakesutil.GetFakeSshClient("fake-user\nfake-devicename", errors.New("fake-error"))
			vm = NewSoftLayerVM(1234567, softLayerClient, sshClient, agentEnvService, logger)

			err := vm.DetachDisk(disk)
			Expect(err).To(HaveOccurred())
		})
	})
})
