package test_helpers_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testhelpers "github.com/maximilien/bosh-softlayer-cpi/test_helpers"
)

var _ = Describe("helper functions for integration tests", func() {
	var (
		cpiTemplate testhelpers.CpiTemplate

		replacementMap map[string]string

		rootTemplatePath string
	)

	BeforeEach(func() {
		replacementMap = map[string]string{
			"ID":           "some ID",
			"DirectorUuid": "some director UUID",
			"Tags":         "some, tags",
		}

		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		rootTemplatePath = filepath.Join(pwd, "..")
	})

	Context("#RunCpi", func() {
		It("/out/cpi to exist and run", func() {
			configPath := filepath.Join(rootTemplatePath, "dev", "config.json")
			payload := `{
							"method": "set_vm_metadata",
							"arguments": [
								"some ID",
								{
									"director": "BOSH Director",
									"deployment": "softlayer",
									"compiling": "nats",
								}
							],
							"context": {
								"director_uuid": "some director UUID"
							}
						}`

			_, err := testhelpers.RunCpi(rootTemplatePath, configPath, payload)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("#CreateTmpConfigPath", func() {
		It("creates a config.json in temp dir with new config template values", func() {
			tmpConfigPath, err := testhelpers.CreateTmpConfigPath(rootTemplatePath, "test_fixtures/cpi_methods/config.json", "some username", "some ApiKey")
			Expect(err).ToNot(HaveOccurred())

			fileInfo, err := os.Stat(tmpConfigPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileInfo.Mode().IsRegular()).To(BeTrue())

			err = os.RemoveAll(tmpConfigPath)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("#GenerateCpiJsonPayload", func() {
		Context("set_vm_metadata CPI method", func() {
			BeforeEach(func() {
				cpiTemplate = testhelpers.CpiTemplate{
					ID:             "fake-id",
					DirectorUuid:   "fake-director-uuid",
					Tag_compiling:  "fake-compiling",
					Tag_deployment: "fake-deployment",
				}
			})

			It("verifies the generated payload", func() {
				payload, err := testhelpers.GenerateCpiJsonPayload("set_vm_metadata", rootTemplatePath, replacementMap)
				Expect(err).ToNot(HaveOccurred())
				Expect(payload).To(MatchJSON(`
					{
						"method": "set_vm_metadata",
						"arguments": [
							"some ID",
							{
								"director": "BOSH Director",
								"deployment": "",
								"compiling": ""
							}
						],
						"context": {
							"director_uuid": "some director UUID"
						}
					}
				`))
			})

			It("fails due to non-existant json template", func() {
				_, err := testhelpers.GenerateCpiJsonPayload("does_not_exist", rootTemplatePath, replacementMap)
				Expect(err).To(HaveOccurred())
			})

		})
	})
})
