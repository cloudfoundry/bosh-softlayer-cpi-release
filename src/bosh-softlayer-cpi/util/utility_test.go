package util_test

import (
	. "github.com/cloudfoundry/bosh-softlayer-cpi/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utility", func() {
	var (
		result string
	)

	Context("#GetOSEnvVariable", func() {
		It("returns the default value if the environment variable is not set", func() {
			result = GetOSEnvVariable("VAR_NAME_NOT_SET", "theDefaultValue")
			Expect(result).To(Equal("theDefaultValue"))
		})
	})

	Context("#SetOSEnvVariable", func() {
		It("returns the expected value set by SetOSEnvVariable", func() {
			err := SetOSEnvVariable("TEST_VARSET", "setByUnitTest")
			Expect(err).ToNot(HaveOccurred())
			result = GetOSEnvVariable("TEST_VARSET", "")
			Expect(result).To(Equal("setByUnitTest"))
		})
	})

})
