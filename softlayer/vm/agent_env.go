package vm

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type AgentEnv struct {
	AgentID string `json:"agent_id"`

	VM VMSpec `json:"vm"`

	Mbus string   `json:"mbus"`
	NTP  []string `json:"ntp"`

	Blobstore BlobstoreSpec `json:"blobstore"`

	Networks Networks `json:"networks"`

	Disks DisksSpec `json:"disks"`

	Env EnvSpec `json:"env"`
}

type VMSpec struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type DisksSpec struct {
	Ephemeral  string         `json:"ephemeral"`
	Persistent PersistentSpec `json:"persistent"`
}

type PersistentSpec map[string]string

type EnvSpec map[string]interface{}

const (
	BlobstoreTypeDav   = "dav"
	BlobstoreTypeLocal = "local"
)

type BlobstoreSpec struct {
	Provider string                 `json:"provider"`
	Options  map[string]interface{} `json:"options"`
}

type DavConfig map[string]interface{}

func NewAgentEnvFromJSON(bytes []byte) (AgentEnv, error) {
	var agentEnv AgentEnv

	err := json.Unmarshal(bytes, &agentEnv)
	if err != nil {
		return agentEnv, bosherr.WrapError(err, "Unmarshalling agent env")
	}

	return agentEnv, nil
}

func NewAgentEnvForVM(agentID, vmCID string, networks Networks, disksSpec DisksSpec, env Environment, agentOptions AgentOptions) AgentEnv {
	networksSpec := Networks{}

	for netName, network := range networks {
		networksSpec[netName] = Network{
			Type: network.Type,

			IP:      network.IP,
			Netmask: network.Netmask,
			Gateway: network.Gateway,

			DNS:           network.DNS,
			Default:       network.Default,
			Preconfigured: true,

			MAC: "",

			CloudProperties: network.CloudProperties,
		}
	}

	agentEnv := AgentEnv{
		AgentID: agentID,

		VM: VMSpec{
			Name: vmCID, // id for name and id
			ID:   vmCID,
		},

		Mbus: agentOptions.Mbus,
		NTP:  agentOptions.NTP,

		Blobstore: BlobstoreSpec{
			Provider: agentOptions.Blobstore.Type,
			Options:  agentOptions.Blobstore.Options,
		},

		Disks: disksSpec,

		Networks: networksSpec,

		// todo deep copy env?
		Env: EnvSpec(env),
	}

	return agentEnv
}

func (ae AgentEnv) AttachPersistentDisk(diskID, path string) AgentEnv {
	spec := PersistentSpec{}

	if ae.Disks.Persistent != nil {
		for k, v := range ae.Disks.Persistent {
			spec[k] = v
		}
	}

	spec[diskID] = path

	ae.Disks.Persistent = spec

	return ae
}

func (ae AgentEnv) DetachPersistentDisk(diskID string) AgentEnv {
	spec := PersistentSpec{}

	if ae.Disks.Persistent != nil {
		for k, v := range ae.Disks.Persistent {
			spec[k] = v
		}
	}

	delete(spec, diskID)

	ae.Disks.Persistent = spec

	return ae
}
