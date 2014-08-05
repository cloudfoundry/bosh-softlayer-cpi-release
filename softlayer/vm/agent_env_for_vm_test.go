package vm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("NewAgentEnvForVM", func() {
	It("returns agent env", func() {
		networks := Networks{
			"fake-net-name": Network{
				Type: "fake-type",

				IP:      "fake-ip",
				Netmask: "fake-netmask",
				Gateway: "fake-gateway",

				DNS:     []string{"fake-dns"},
				Default: []string{"fake-default"},

				CloudProperties: map[string]interface{}{
					"fake-cp-key": "fake-cp-value",
				},
			},
		}

		env := Environment{"fake-env-key": "fake-env-value"}

		agentOptions := AgentOptions{
			Mbus: "fake-mbus",
			NTP:  []string{"fake-ntp"},

			Blobstore: BlobstoreOptions{
				Type: "fake-blobstore-type",
				Options: map[string]interface{}{
					"fake-blobstore-key": "fake-blobstore-value",
				},
			},
		}

		agentEnv := NewAgentEnvForVM(
			"fake-agent-id",
			"fake-vm-id",
			networks,
			env,
			agentOptions,
		)

		expectedAgentEnv := AgentEnv{
			AgentID: "fake-agent-id",

			VM: VMSpec{
				Name: "fake-vm-id", // id for name and id
				ID:   "fake-vm-id",
			},

			Mbus: "fake-mbus",

			NTP: []string{"fake-ntp"},

			Blobstore: BlobstoreSpec{
				Provider: "fake-blobstore-type",
				Options: map[string]interface{}{
					"fake-blobstore-key": "fake-blobstore-value",
				},
			},

			Networks: NetworksSpec{
				"fake-net-name": NetworkSpec{
					Type: "fake-type",

					IP:      "fake-ip",
					Netmask: "fake-netmask",
					Gateway: "fake-gateway",

					DNS:     []string{"fake-dns"},
					Default: []string{"fake-default"},

					MAC: "",

					CloudProperties: map[string]interface{}{
						"fake-cp-key": "fake-cp-value",
					},
				},
			},

			Env: map[string]interface{}{
				"fake-env-key": "fake-env-value",
			},
		}

		Expect(agentEnv).To(Equal(expectedAgentEnv))
	})
})
