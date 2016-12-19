package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

var _ = Describe("NewAgentEnvForVM", func() {
	It("returns agent env", func() {
		networks := Networks{
			"fake-net-name": Network{
				Type: "fake-type",

				IP:      "fake-ip",
				Netmask: "fake-netmask",
				Gateway: "fake-gateway",

				DNS:           []string{"fake-dns"},
				Default:       []string{"fake-default"},
				Preconfigured: true,

				CloudProperties: map[string]interface{}{
					"fake-cp-key": "fake-cp-value",
				},
			},
		}

		disks := DisksSpec{Ephemeral: "/dev/xvdc"}

		env := Environment{"fake-env-key": "fake-env-value"}

		agentOptions := AgentOptions{
			Mbus: "fake-mbus",
			NTP:  []string{"fake-ntp"},

			Blobstore: BlobstoreOptions{
				Provider: "fake-blobstore-type",
				Options: map[string]interface{}{
					"fake-blobstore-key": "fake-blobstore-value",
				},
			},
		}

		agentEnv := NewAgentEnvForVM(
			"fake-agent-id",
			"fake-vm-id",
			networks,
			disks,
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

			Networks: Networks{
				"fake-net-name": Network{
					Type: "fake-type",

					IP:      "fake-ip",
					Netmask: "fake-netmask",
					Gateway: "fake-gateway",

					DNS:           []string{"fake-dns"},
					Default:       []string{"fake-default"},
					Preconfigured: true,

					MAC: "",

					CloudProperties: map[string]interface{}{
						"fake-cp-key": "fake-cp-value",
					},
				},
			},

			Disks: DisksSpec{
				Ephemeral: "/dev/xvdc",
			},

			Env: map[string]interface{}{
				"fake-env-key": "fake-env-value",
			},
		}

		Expect(agentEnv).To(Equal(expectedAgentEnv))
	})
})
