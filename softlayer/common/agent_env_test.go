package common_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
)

var _ = Describe("AgentEnv", func() {
	Describe("AttachPersistentDisk", func() {
		It("sets persistent disk path for given disk id", func() {
			agentEnv := AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-other-disk-id": "fake-other-disk-path",
					},
				},
			}

			newAgentEnv := agentEnv.AttachPersistentDisk("fake-disk-id", "fake-disk-path")

			Expect(newAgentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-other-disk-id": "fake-other-disk-path",
						"fake-disk-id":       "fake-disk-path",
					},
				},
			}))

			// keeps original agent env not modified
			Expect(agentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-other-disk-id": "fake-other-disk-path",
					},
				},
			}))
		})

		It("sets persistent disk path for given disk id on an empty agent env", func() {
			agentEnv := AgentEnv{}

			newAgentEnv := agentEnv.AttachPersistentDisk("fake-disk-id", "fake-disk-path")

			Expect(newAgentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-disk-id": "fake-disk-path",
					},
				},
			}))

			// keeps original agent env not modified
			Expect(agentEnv).To(Equal(AgentEnv{}))
		})

		It("overwrites persistent disk path for given disk id", func() {
			agentEnv := AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-disk-id": "fake-old-disk-path",
					},
				},
			}

			newAgentEnv := agentEnv.AttachPersistentDisk("fake-disk-id", "fake-new-disk-path")

			Expect(newAgentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-disk-id": "fake-new-disk-path",
					},
				},
			}))

			// keeps original agent env not modified
			Expect(agentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-disk-id": "fake-old-disk-path",
					},
				},
			}))
		})
	})

	Describe("DetachPersistentDisk", func() {
		It("unsets persistent disk path if previously set", func() {
			agentEnv := AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-disk-id": "fake-disk-path",
					},
				},
			}

			newAgentEnv := agentEnv.DetachPersistentDisk("fake-disk-id")

			Expect(newAgentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{},
				},
			}))

			// keeps original agent env not modified
			Expect(agentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-disk-id": "fake-disk-path",
					},
				},
			}))
		})

		It("unsets persistent disk path on an empty agent env", func() {
			agentEnv := AgentEnv{}

			newAgentEnv := agentEnv.DetachPersistentDisk("fake-disk-id")

			Expect(newAgentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{},
				},
			}))

			// keeps original agent env not modified
			Expect(agentEnv).To(Equal(AgentEnv{}))
		})

		It("does not change anything if persistent disk was not set", func() {
			agentEnv := AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-other-disk-id": "fake-other-disk-path",
					},
				},
			}

			newAgentEnv := agentEnv.DetachPersistentDisk("fake-disk-id")

			Expect(newAgentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-other-disk-id": "fake-other-disk-path",
					},
				},
			}))

			// keeps original agent env not modified
			Expect(agentEnv).To(Equal(AgentEnv{
				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-other-disk-id": "fake-other-disk-path",
					},
				},
			}))
		})
	})

	Describe("MarshalJSON", func() {
		It("returns JSON encoded agent env", func() {
			agentEnv := AgentEnv{
				AgentID: "fake-agent-id",

				VM: VMSpec{
					Name: "fake-vm-name",
					ID:   "fake-vm-id",
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

						MAC: "fake-mac",

						CloudProperties: map[string]interface{}{
							"fake-cp-key": "fake-cp-value",
						},
					},
				},

				Disks: DisksSpec{
					Persistent: PersistentSpec{
						"fake-persistent-id": "fake-persistent-path",
					},
				},

				Env: EnvSpec{
					"fake-env-key": "fake-env-value",
				},
			}

			jsonBytes, err := json.Marshal(agentEnv)
			Expect(err).ToNot(HaveOccurred())

			outAgentEnv, err := NewAgentEnvFromJSON(jsonBytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(outAgentEnv).To(Equal(agentEnv))
		})

		It("returns error if agent env cannot be JSON encoded", func() {
			agentEnv := AgentEnv{
				Networks: Networks{
					"fake-net-name": Network{
						CloudProperties: map[string]interface{}{
							"fake-key": NonJSONMarshable{},
						},
					},
				},
			}

			_, err := json.Marshal(agentEnv)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-marshal-err"))
		})
	})
})
