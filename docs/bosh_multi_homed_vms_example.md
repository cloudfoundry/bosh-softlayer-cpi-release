# BOSH Multi-homed VMs example

A very simple BOSH SoftLayer [Multi-homed VMs](https://bosh.io/docs/networks.html#multi-homed) example.

### 1. Upload the Dummy BOSH release
```
bosh2 upload-release https://github.com/pivotal-cf-experimental/dummy-boshrelease/releases/download/v2/dummy-2.tgz
```

### 2. Update the BOSH director's cloud-config

```
cat >cloud-config.yml << EOF
azs:
- cloud_properties:
    datacenter: lon02
  name: lon02

compilation:
  az: lon02
  network: private
  reuse_compilation_vms: true
  vm_type: compilation
  workers: 8

vm_types:
- cloud_properties:
    max_network_speed: 1000
    ephemeral_disk_size: 100
    hourly_billing_flag: true
    local_disk_flag: true
    memory: 8192
    cpu: 4
    hostname_prefix: ed-stage0-worker
  name: compilation
- cloud_properties:
    max_network_speed: 1000
    ephemeral_disk_size: 100
    hourly_billing_flag: false
    local_disk_flag: true
    memory: 8192
    cpu: 4
    hostname_prefix: ed-stage0-core
    dedicated_account_host_only_flag: false
  name: coreNode

networks:
- name: public
  subnets:
  - az: lon02
    cloud_properties:
      vlan_ids:
      - 524956
    dns:
    - 10.165.232.98
    - 8.8.8.8
    - 10.0.80.11
    - 10.0.80.12
  type: dynamic
- name: private
  subnets:
  - az: lon02
    cloud_properties:
      vlan_ids:
      - 524954
    dns:
    - 10.165.232.98
    - 8.8.8.8
    - 10.0.80.11
    - 10.0.80.12
  type: dynamic
EOF
bosh2 update-cloud-config cloud-config.yml
```

### 3. Deploy dummy through the sample manifest
Configures `default` property for specific `addressable`. `addressable` can be used to specify which IP address other jobs see.

```
cat >dummy.yml << EOF
---
name: dummy

releases:
- name: dummy
  version: latest

stemcells:
- alias: ubuntu
  os: ubuntu-trusty
  version: latest

instance_groups:
- azs:
  - lon02
  name: dummy
  instances: 1
  vm_type: coreNode
  stemcell: ubuntu
  networks:
  - { name: public, default: [dns, gateway] }
  - { name: private, default: [addressable] }
  jobs:
  - name: dummy
    release: dummy

update:
  canaries: 1
  max_in_flight: 6
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000
EOF
bosh2 deploy -d dummy dummy.yml
```

### 4. Verify deployment
```
bosh2 vms --dns
Using environment 'bosh-director-automation-cf.softlayer.com' as client 'admin'

Task 12. Done

Deployment 'dummy'

Instance                                    Process State  AZ     IPs             VM CID    VM Type   DNS A Records
dummy/149c540d-0288-40bf-b4a3-edbfc65b51da  running        lon02  169.50.10.10    XXXXXXXX  coreNode  XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX.dummy.public.dummy.bosh
                                                                  10.10.10.10                         XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX.dummy.private.dummy.bosh
                                                                                                      0.dummy.public.dummy.bosh
                                                                                                      0.dummy.private.dummy.bosh

1 vms

Succeeded
```
