package action_test

import (
	. "bosh-softlayer-cpi/action"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConcreteFactoryOptions", func() {
	var (
		cloudProps VMCloudProperties
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			cloudProps = VMCloudProperties{
				HostnamePrefix:    "fake-hostname",
				Domain:            "fake-domain.com",
				Cpu:               2,
				Memory:            2048,
				MaxNetworkSpeed:   100,
				Datacenter:        "fake-datacenter",
				SshKey:            32345678,
				DeployedByBoshCLI: true,
			}
		})

		It("does not return error if all fields are valid", func() {
			err := cloudProps.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if HostnamePrefix is not be set", func() {
			cloudProps = VMCloudProperties{
				Domain:            "fake-domain.com",
				Cpu:               2,
				Memory:            2048,
				MaxNetworkSpeed:   100,
				Datacenter:        "fake-datacenter",
				SshKey:            32345678,
				DeployedByBoshCLI: true,
			}

			err := cloudProps.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("The property 'hostname_prefix' must be set to create an instance"))
		})

		It("Add 'Domain' property if Domain is not be set", func() {
			cloudProps = VMCloudProperties{
				HostnamePrefix:    "fake-hostname",
				Cpu:               2,
				Memory:            2048,
				MaxNetworkSpeed:   100,
				Datacenter:        "fake-datacenter",
				SshKey:            32345678,
				DeployedByBoshCLI: true,
			}

			err := cloudProps.Validate()
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudProps.Domain).To(Equal("softlayer.com"))
		})

		It("returns error if Datacenter is not be set", func() {
			cloudProps = VMCloudProperties{
				HostnamePrefix:    "fake-hostname",
				Domain:            "fake-domain.com",
				Cpu:               2,
				Memory:            2048,
				MaxNetworkSpeed:   100,
				SshKey:            32345678,
				DeployedByBoshCLI: true,
			}

			err := cloudProps.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("The property 'datacenter' must be set to create an instance"))
		})
	})
})
