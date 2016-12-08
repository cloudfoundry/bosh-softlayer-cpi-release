package clients

// /info

type DataInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InfoResponse struct {
	Status int      `json:"status"`
	Data   DataInfo `json:"data"`
}

// /bms
type BaremetalInfo struct {
	Id                 int      `json:"id"`
	Hostname           string   `json:"hostname"`
	Private_ip_address string   `json:"private_ip_address"`
	Public_ip_address  string   `json:"public_ip_address"`
	Tags               []string `json:"tags"`
	Memory             int      `json:"memory"`
	Cpu                int      `json:"cpu"`
	Provision_date     string   `json:"provision_date"`
}

// /baremetal/spec/${server_name}/${stemcell}/${netboot_image}
type ProvisioningBaremetalInfo struct {
	VmNamePrefix     string `json:"vmNamePrefix,omitempty"`
	Bm_stemcell      string `json:"bm_stemcell,omitempty"`
	Bm_netboot_image string `json:"bm_netboot_image,omitempty"`
}

type BmsResponse struct {
	Status int             `json:"Status"`
	Data   []BaremetalInfo `json:"data"`
}

// /sl/packages

type Package struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type DataPackage struct {
	Packages []Package `json:"packages"`
}

type SlPackagesResponse struct {
	Status int         `json:"status"`
	Data   DataPackage `json:"data"`
}

// /sl/${package_id}/options
type Option struct {
	Id          int    `json:"id"`
	Description string `json:"description"`
}

type Category struct {
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	Options  []Option `json:"options"`
	Required bool     `json:"required"`
}

type DataPackageOptions struct {
	Category   []Category `json:"categories"`
	Datacenter []string   `json:"datacenters"`
}

type SlPackageOptionsResponse struct {
	Status int                `json:"status"`
	Data   DataPackageOptions `json:"data"`
}

// /stemcells

type StemcellsResponse struct {
	Status   int      `json:"status"`
	Stemcell []string `json:"data"`
}

// /tasks?latest= (default 50)

type Task struct {
	Id          int    `json:"id"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	Status      string `json:"status"`
	EndTime     string `json:"end_time"`
}

type TasksResponse struct {
	Status int    `json:"status"`
	Data   []Task `json:"data"`
}

// /task/${task_id}/txt}" (default event)

type TaskOutputResponse struct {
	Status int      `json:"status"`
	Data   []string `json:"data"`
}

type TaskJsonResponse struct {
	Status int                    `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// /baremetal/${serverId}/${status}

type UpdateStateResponse struct {
	Status int `json:"status"`
}

// /login/${username}/${password}

type LoginResponse struct {
	Status int `json:"status"`
}

// //baremetals (dry_run: optional)

type TaskInfo struct {
	TaskId int `json:"task_id"`
}

type CreateBaremetalsResponse struct {
	Status int      `json:"status"`
	Data   TaskInfo `json:"data"`
}

type ServerSpec struct {
	Cores         int  `yaml:"cores" json:"cores,omitempty"`
	Memory        int  `yaml:"memory" json:"memory,omitempty"`
	MaxPortSpeed  int  `yaml:"max_port_speed" json:"max_port_speed,omitempty"`
	PublicVlanId  int  `yaml:"public_vlan_id" json:"public_vlan_id,omitempty"`
	PrivateVlanId int  `yaml:"private_vlan_id" json:"private_vlan_id,omitempty"`
	Hourly        bool `yaml:"hourly" json:"hourly"`
}

type CloudProperty struct {
	BoshIP         string     `yaml:"bosh_ip" json:"bosh_ip"`
	Datacenter     string     `yaml:"datacenter" json:"datacenter"`
	Domain         string     `yaml:"domain" json:"domain"`
	NamePrefix     string     `yaml:"name_prefix" json:"name_prefix"`
	ServerSpec     ServerSpec `yaml:"server_spec" json:"server_spec"`
	Baremetal      bool       `yaml:"baremetal" json:"baremetal"`
	BmStemcell     string     `yaml:"bm_stemcell" json:"bm_stemcell,omitempty"`
	BmNetbootImage string     `yaml:"bm_netboot_image" json:"bm_netboot_image,omitempty"`
	Size           int        `yaml:"size" json:"size"`
}

type CreateBaremetalsParameters struct {
	Parameters CreateBaremetalsInfo `json:"parameters"`
}

type CreateBaremetalsInfo struct {
	BaremetalSpecs []CloudProperty `json:"baremetal_specs"`
	Deployment     string          `json:"deployment"`
}

type Stemcell struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Resource struct {
	Name          string        `yaml:"name"`
	Network       string        `yaml:"network"`
	Size          int           `yaml:"size"`
	Stemcell      Stemcell      `yaml:"stemcell"`
	CloudProperty CloudProperty `yaml:"cloud_properties"`
}

// deployment
type Deployment struct {
	Name          string     `yaml:"name"`
	ResourcePools []Resource `yaml:"resource_pools"`
}
