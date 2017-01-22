#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param BASE_OS
check_param BAT_VCAP_PASSWORD
check_param SL_DATACENTER
check_param SL_VLAN_PUBLIC
check_param SL_VLAN_PRIVATE
check_param SL_VM_NAME_PREFIX
check_param SL_VM_DOMAIN
check_param BAT_DIRECTOR_IP


DIRECTOR=$BAT_DIRECTOR_IP

source /etc/profile.d/chruby.sh
chruby 2.2.4

echo "DirectorIP =" $DIRECTOR

export BAT_DIRECTOR=$DIRECTOR
export BAT_DNS_HOST=$DIRECTOR
export BAT_STEMCELL=`echo $PWD/stemcell/*.tgz`
export BAT_DEPLOYMENT_SPEC="${PWD}/${BASE_OS}-bats-config.yml"
export BAT_VCAP_PASSWORD=$BAT_VCAP_PASSWORD
export BAT_INFRASTRUCTURE=softlayer
export BAT_NETWORKING=dynamic
export BAT_DEBUG_MODE=true

STEMCELL_VERSION=$(cat stemcell-version/number | cut -f1 -d.)

bosh -n target $BAT_DIRECTOR
echo Using This version of bosh:
bosh --version
cat > "${BAT_DEPLOYMENT_SPEC}" <<EOF
---
cpi: softlayer
manifest_template_path: $(echo `pwd`/softlayer.yml.erb)
properties:
  uuid: $(bosh status --uuid)
  stemcell:
    name: bosh-softlayer-xen-ubuntu-trusty-go_agent
    version: latest
  cloud_properties:
    bosh_ip: $BAT_DIRECTOR
    public_vlan_id: $SL_VLAN_PUBLIC
    private_vlan_id: $SL_VLAN_PRIVATE
    vm_name_prefix: $SL_VM_NAME_PREFIX
    data_center: $SL_DATACENTER
    domain: $SL_VM_DOMAIN
  pool_size: 1
  instances: 1
  password: "\$6\$3n/Y5RP0\$Jr1nLxatojY9Wlqduzwh66w8KmYxjoj9vzI62n3Mmstd5mNVnm0SS1N0YizKOTlJCY5R/DFmeWgbkrqHIMGd51"
  networks:
  - name: default
    type: dynamic
    dns:
    - $BAT_DNS_HOST
  batlight:
    missing: nope
EOF
cat > softlayer.yml.erb <<EOF
---
name: <%= properties.name || "bat" %>
director_uuid: <%= properties.uuid %>

releases:
  - name: bat
    version: <%= properties.release || "latest" %>

compilation:
  workers: 1
  network: default
  reuse_compilation_vms: true
  cloud_properties:
    Bosh_ip: <%= properties.cloud_properties.bosh_ip %>
    StartCpus: 4
    MaxMemory: 8192
    EphemeralDiskSize: 100
    Datacenter:
      Name: <%= properties.cloud_properties.data_center %>
    VmNamePrefix: bat-softlayer-worker-
    Domain: softlayer.com
    HourlyBillingFlag: true
    PrimaryNetworkComponent:
      NetworkVlan:
        Id: <%= properties.cloud_properties.public_vlan_id %>
    PrimaryBackendNetworkComponent:
      NetworkVlan:
        Id: <%= properties.cloud_properties.private_vlan_id %>
    NetworkComponents:
    - MaxSpeed: 1000

update:
  canaries: <%= properties.canaries || 1 %>
  canary_watch_time: 3000-90000
  update_watch_time: 3000-90000
  max_in_flight: <%= properties.max_in_flight || 1 %>
  serial: true

networks:
<% properties.networks.each do |network| %>
- name: default
  type: dynamic
  dns: <%= p('dns').inspect %>
  cloud_properties:
    security_groups:
    - default
    - cf
<% end %>

resource_pools:
  - name: common
    network: default
    size: <%= properties.pool_size %>
    stemcell:
      name: <%= properties.stemcell.name %>
      version: '<%= properties.stemcell.version %>'
    cloud_properties:
      Bosh_ip: <%= properties.cloud_properties.bosh_ip %>
      StartCpus: 4
      MaxMemory: 8192
      EphemeralDiskSize: 100
      Datacenter:
        Name: <%= properties.cloud_properties.data_center %>
      VmNamePrefix: <%= properties.cloud_properties.vm_name_prefix %>
      Domain: <%= properties.cloud_properties.domain %>
      HourlyBillingFlag: true
      PrimaryNetworkComponent:
        NetworkVlan:
          Id: <%= properties.cloud_properties.public_vlan_id %>
      PrimaryBackendNetworkComponent:
        NetworkVlan:
          Id: <%= properties.cloud_properties.private_vlan_id %>
      NetworkComponents:
      - MaxSpeed: 1000
    <% if properties.password %>
    env:
      bosh:
        password: <%= properties.password %>
        keep_root_password: true
    <% end %>

jobs:
  - name: <%= properties.job || "batlight" %>
    templates: <% (properties.templates || ["batlight"]).each do |template| %>
    - name: <%= template %>
    <% end %>
    instances: <%= properties.instances %>
    resource_pool: common
    <% if properties.persistent_disk %>
    persistent_disk: <%= properties.persistent_disk %>
    <% end %>
    networks:
      <% properties.job_networks.each_with_index do |network, i| %>
      - name: default
        <% if i == 0 %>
        default: [dns, gateway]
        <% end %>
      <% end %>

properties:
  batlight:
    <% if properties.batlight.fail %>
    fail: <%= properties.batlight.fail %>
    <% end %>
    <% if properties.batlight.missing %>
    missing: <%= properties.batlight.missing %>
    <% end %>
    <% if properties.batlight.drain_type %>
    drain_type: <%= properties.batlight.drain_type %>
    <% end %>
EOF

cd bats
./write_gemfile
bundle install
bundle exec rspec spec