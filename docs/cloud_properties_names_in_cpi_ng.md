# Cloud properties names in CPI NG

## manifest change  

This topic describes properties of deployment by comparing with the old SoftLayer CPI.

### AZs

AZs support the Director to eliminate and/or simplify manual configuration for balancing VMs across AZs and IP address management. Once AZs are defined, deployment jobs can be placed into one or more AZs.

AZs schema:  
  * azs [Array, required]: List of AZs.
  * name [String, required]: Name of an AZ within the Director.
  * cloud_properties [Hash, optional]: Describes any IaaS-specific properties needed to associated with AZ; for most IaaSes, some data here is actually required. See CPI Specific cloud_properties below. Example: availability_zone. Default is {} (empty Hash).
    - datacenter [String&lt;String&gt;, required]: Name of the datacenter. Example: `lon02`.

_Note that IaaS specific cloud properties related to AZs should now be only placed under azs. Make sure to remove them from resource_pools/vm_types’ cloud properties._

sample manifest for current softlayer cpi:
```yaml
azs:
- name: z1
  cloud_properties:
    Datacenter: { Name: lon02 }
```

sample manifest for new softlayer cpi:  
```yaml
azs:
- name: z1
  cloud_properties:
    datacenter: lon02 
```

### Networks

There are three different network types: `manual`, `dynamic`, and `vip`. And softlayer-cpi does not support `vip` type at present.

Manual Networks schema:
  * name [String, required]: Name used to reference this network configuration
  * type [String, required]: Value should be manual
  * subnets [Array, required]: Lists subnets in this network
    * range [String, required]: Subnet IP range that includes all IPs from this subnet
    * gateway [String, required]: Subnet gateway IP
    * dns [Array, optional]: DNS IP addresses for this subnet
    * reserved [Array, optional]: Array of reserved IPs and/or IP ranges. BOSH does not assign IPs from this range to any VM
    * static [Array, optional]: Array of static IPs and/or IP ranges. BOSH assigns IPs from this range to jobs requesting static IPs. Only IPs specified here can be used for static IP reservations.
    * az [String, optional]: AZ associated with this subnet (should only be used when using first class AZs). Example: z1. Available in v241+.
    * azs [Array, optional]: List of AZs associated with this subnet (should only be used when using first class AZs). Example: [z1, z2]. Available in v241+.
    * cloud_properties [Hash, optional]: Describes any IaaS-specific properties for the subnet. Default is {} (empty Hash).
      - vlan_ids [Array&lt;String&gt;, required]: A list of the [SoftLayer Network Vlan](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Vlan) id that the CPI will use when creating the instance (at lest set one private network). Example: `524954`.

sample manifest of static network for current softlayer cpi:
```yaml
networks:
- name: default
  type: manual
  subnets:
  - range: 10.112.166.128/26
    gateway: 10.112.166.129
    dns: [8.8.8.8, 10.0.80.11, 10.0.80.12]
    static: [10.112.166.131]
    az: z1
    cloud_properties:
      PrimaryNetworkComponent:
        NetworkVlan:
          Id: 524956
      PrimaryBackendNetworkComponent:
        NetworkVlan:
          Id: 524954
```

sample manifest of static network for new softlayer cpi:  
```yaml
networks:
- name: default
  type: manual
  subnets:
  - range: 10.112.166.128/26
    gateway: 10.112.166.129
    dns: [8.8.8.8, 10.0.80.11, 10.0.80.12]
    static: [10.112.166.131]
    az: z1
    cloud_properties:
      vlan_ids: [524956, 524954]
```

Dynamic networks schema:  
  * name [String, required]: Name used to reference this network configuration
  * type [String, required]: Value should be dynamic
  * dns [Array, optional]: DNS IP addresses for this network
  * cloud_properties [Hash, optional]: Describes any IaaS-specific properties for the network. Default is {} (empty Hash).
      - vlan_ids [Array&lt;String&gt;, required]: A list of the [SoftLayer Network Vlan](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Network_Vlan) id that the CPI will use when creating the instance (at lest set one private network). Example: `524954`.

sample manifest of dynamic network for current softlayer cpi:
```yaml
networks
- name: dynamic
  type: dynamic
  dns: [8.8.8.8, 10.0.80.11, 10.0.80.12]
  cloud_properties:
    PrimaryNetworkComponent:
      NetworkVlan:
        Id: 524956
    PrimaryBackendNetworkComponent:
      NetworkVlan:
        Id: 524954
```

sample manifest of dynamic network for new softlayer cpi:  
```yaml
networks
- name: dynamic
  type: dynamic
  cloud_properties:
    vlan_ids: [((sl_vlan_public_id)), ((sl_vlan_private_id))]
  dns: [8.8.8.8, 10.0.80.11, 10.0.80.12]
```

### VM Types

VM type is a named Virtual Machine size configuration in the cloud config.
Resource pool is collections of VMs created from the same stemcell, with the same configuration, in a deployment.

Vm_types schema:  
  * vm_types [Array, required]: Specifies the VM types available to deployments. At least one should be specified.
    * name [String, required]: A unique name used to identify and reference the VM type
    * cloud_properties [Hash, optional]: Describes any IaaS-specific properties needed to create VMs; for most IaaSes, some data here is actually required. See CPI Specific cloud_properties below. Example: instance_type: m3.medium. Default is {} (empty Hash).
      - hostname_prefix** [String, required]: The hostname of the [SoftLayer Virtual Guest](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest) the CPI will use when creating the instance. Please note that, for bosh director, this property is the full hostname, and for other instances in deployments, a timestamp will be appended to the property value to make the hostname. Example: `bosh-softlayer`.
      - domain** [String, required]: The domain name of the [SoftLayer Virtual Guest](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest) the CPI will use when creating the instance. Example: `softlayer.com`.
      - cpu** [Integer, required]: Number of CPUs ([SoftLayer Virtual Guest](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest)) the CPI will use when creating the instance. Example: `4`.
      - memory** [Integer, required]: Memory(in Mb) ([SoftLayer Virtual Guest](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest)) the CPI will use when creating the instance. Example: `8192`.
      - max_network_speed** [Integer, optional]: Max speed of networkComponents SoftLayer Virtual Guest](https://sldn.softlayer.com/reference/datatypes/SoftLayer_Virtual_Guest) the CPI will use when creating the instance. Default is `1000`.
      - ephemeral_disk_size** [Integer, optional]: Ephemeral disk size(in Gb) the CPI will use when creating the instance. Example: `100`.
      - hourly_billing_flag** [Boolean, optional]: If the instance is hourly billing. Default is `false`.
      - local_disk_flag** [Boolean, optional]: If the instance has at least one disk which is local to the host it runs on. Default is `false`.
      - dedicated_account_host_only_flag** [Boolean, optional]: If the instance is to run on hosts that only have guests from the same account. Conflicts with `dedicated_host_id`. Default is `false`.
      - dedicated_host_id** [Integer, optional]: Specifies dedicated host for the instance by its id. Conflicts with `dedicated_acc_host_only_flag`. Default is '0'.
      - deployed_by_boshcli** [Boolean, optional]: If the instance is deployed by bosh-cli. Default is `false`.

sample manifest of current softlayer cpi:
```yaml
vm_types:
- name: default
  cloud_properties:
    bosh_ip: ((internal_ip))
    startCpus:  4
    maxMemory:  8192
    ephemeralDiskSize: 100
    hourlyBillingFlag: true
    vmNamePrefix: manifest-sample
    domain: sofltayer.com
```

sample manifest for new softlayer cpi:  
```yaml
vm_types:
- name: default
  cloud_properties:
    cpu:  4
    memory:  8192
    ephemeral_disk_size: 100
    hourly_billing_flag: true
    hostname_prefix: manifest-sample
    domain: sofltayer.com
```

### Disk types

Disk Type (previously known as Disk Pool) is a named disk configuration specified in the cloud config.

* disk_types [Array, required]: Specifies the disk types available to deployments. At least one should be specified.
    * name [String, required]: A unique name used to identify and reference the disk type
    * disk_size [Integer, required]: Specifies the disk size. disk_size must be a positive integer. BOSH creates a persistent disk of that size in megabytes and attaches it to each job instance VM.
    * cloud_properties [Hash, optional]: Describes any IaaS-specific properties needed to create disks. Examples: type, iops. Default is {} (empty Hash).
      - iops** [Integer, optional]: Input/output operations per second (IOPS) value. Example: `1000`. If it's not set, a medium IOPS value of the specified disk size will be chosen.
      - snapshot_space** [Boolean, optional]: The size of snapshot space of the disk. Example: `20`.

sample manifest of current softlayer cpi:
```yaml
disk_types:
- name: disks
  disk_size: 100_000
  cloud_properties:
    Iops: 3000
    UseHourlyPricing: true
```

sample manifest for new softlayer cpi:  
```yaml
disk_types:
- name: disks
  disk_size: 100_000
  cloud_properties:
    iops: 3000
    snapshot_space: 20

```
    
### Typical sample: The director deployment manifest

current softlayer cpi:
```yaml
cloud_provider:
  cert:
    ca: ...
    certificate: ...
    private_key: ...
  mbus: https://mbus:lvwczko3gux720lijapg@10.112.166.130:6868
  properties:
    agent:
      mbus: https://mbus:lvwczko3gux720lijapg@0.0.0.0:6868
    blobstore:
      path: /var/vcap/micro_bosh/data/cache
      provider: local
    ntp:
    - time1.google.com
    - time2.google.com
    - time3.google.com
    - time4.google.com
    softlayer:
      apiKey: ...
      username: ...
  template:
    name: softlayer_cpi
    release: bosh-softlayer-cpi
disk_pools:
- disk_size: 200000
  name: disks
instance_groups: ...
name: bosh
networks:
- name: default
  subnets:
  - cloud_properties:
      PrimaryNetworkComponent:
        NetworkVlan:
          Id: 1234567
    dns:
    - 8.8.8.8
    gateway: 10.112.166.129
    range: 10.112.166.128/26
    static:
    - 10.112.166.130
  type: manual
- cloud_properties:
    PrimaryBackendNetworkComponent:
      NetworkVlan:
        Id: 1234567
    PrimaryNetworkComponent:
      NetworkVlan:
        Id: 1234568
  dns:
  - 8.8.8.8
  - 10.0.80.11
  - 10.0.80.12
  name: dynamic
  type: dynamic
releases: ...
resource_pools:
- cloud_properties:
    datacenter:
      name: lon02
    deployedByBoshcli: true
    domain: fake-domain.com
    ephemeralDiskSize: 100
    hourlyBillingFlag: true
    maxMemory: 8192
    networkComponents:
    - maxSpeed: 100
    startCpus: 4
    vmNamePrefix: fakie-prefix
  env:
    bosh: ...
  name: vms
  network: dynamic
  stemcell: ...
variables: []
```

new softlayer cpi: 
```yaml
cloud_provider:
  cert:
    ca: ...
    certificate: ...
    private_key: ...
  mbus: https://mbus:511kcx7pje314j5pk8dt@10.112.166.130:6868
  properties:
    agent:
      mbus: https://mbus:511kcx7pje314j5pk8dt@0.0.0.0:6868
    blobstore:
      path: /var/vcap/micro_bosh/data/cache
      provider: local
    ntp:
    - time1.google.com
    - time2.google.com
    - time3.google.com
    - time4.google.com
    softlayer:
      api_key: ...
      username: ...
  ssh_tunnel:
    host: director-sl-cpi-shadow.softlayer.com
    port: 22
    private_key: ...
    user: root
  template:
    name: softlayer_cpi
    release: bosh-softlayer-cpi
disk_pools:
- cloud_properties:
    iops: 3000
    snapshot_space: 20
  disk_size: 200000
  name: disks
instance_groups: ...
name: bosh
networks:
- name: default
  subnets:
  - cloud_properties:
      vlan_ids:
      - 1234567
    dns:
    - 8.8.8.8
    gateway: 10.112.166.129
    range: 10.112.166.128/26
    static:
    - 10.112.166.130
  type: manual
- cloud_properties:
    vlan_ids:
    - 1234567
    - 1234568
  dns:
  - 8.8.8.8
  - 10.0.80.11
  - 10.0.80.12
  name: dynamic
  type: dynamic
releases: ...
resource_pools:
- cloud_properties:
    datacenter: lon02
    deployed_by_boshcli: true
    domain: fake-domain.com
    ephemeral_disk_size: 100
    hourly_billing_flag: true
    memory: 8192
    max_network_speed: 100
    cpu: 4
    hostname_prefix: fake-prefix
  env:
    bosh: ...
  name: vms
  network: default
  stemcell: ...
variables: []
```
