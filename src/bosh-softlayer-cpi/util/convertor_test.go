package util_test

import (
	bscutil "bosh-softlayer-cpi/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Convertor", func() {

	Context("Convert JSON []bytes", func() {
		It("Convert camel case keys to snake case keys ins JSON string", func() {
			jsonCamel := []byte(`
			{
				"vmNamePrefix":"bosh-",
				"domain":"fake-domain.com",
				"startCpus":4,
				"maxMemory":2048
			}`)

			jsonSnake := []byte(`
			{
				"vm_name_prefix":"bosh-",
				"domain":"fake-domain.com",
				"start_cpus":4,
				"max_memory":2048
			}`)
			convertedJSON := bscutil.ConvertJSONKeyCase(jsonCamel)
			Expect(convertedJSON).To(MatchJSON(jsonSnake))
		})

		It("Convert camel case keys to snake case keys in nesting JSON string", func() {
			jsonCamel := []byte(`
			{
				"vmNamePrefix":"bosh-",
				"domain":"fake-domain.com",
				"startCpus":4,
				"primaryBackendNetworkComponent": {
					"networkVlan": {
					  "id": 524954
					}
				},
				"maxMemory":2048
			}`)

			jsonSnake := []byte(`
			{
				"vm_name_prefix":"bosh-",
				"domain":"fake-domain.com",
				"start_cpus":4,
				"primary_backend_network_component": {
					"network_vlan": {
					  "id": 524954
					}
				},
				"max_memory":2048
			}`)
			convertedJSON := bscutil.ConvertJSONKeyCase(jsonCamel)
			Expect(convertedJSON).To(MatchJSON(jsonSnake))
		})

		It("Convert camel case keys to snake case keys in mixed JSON string", func() {
			jsonCamel := []byte(`
			{
				"vmNamePrefix":"bosh-",
				"domain":"fake-domain.com",
				"start_cpus":4,
				"primaryBackendNetworkComponent": {
					"networkVlan": {
					  "id": 524954
					}
				},
				"max_memory":2048
			}`)

			jsonSnake := []byte(`
			{
				"vm_name_prefix":"bosh-",
				"domain":"fake-domain.com",
				"start_cpus":4,
				"primary_backend_network_component": {
					"network_vlan": {
					  "id": 524954
					}
				},
				"max_memory":2048
			}`)
			convertedJSON := bscutil.ConvertJSONKeyCase(jsonCamel)
			Expect(convertedJSON).To(MatchJSON(jsonSnake))
		})
	})
})
