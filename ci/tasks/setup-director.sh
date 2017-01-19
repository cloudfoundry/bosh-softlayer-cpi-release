#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param BASE_OS
check_param SL_USERNAME
check_param SL_API_KEY
check_param SL_VM_NAME_PREFIX
check_param SL_VM_DOMAIN
check_param SL_DATACENTER
check_param SL_VLAN_PUBLIC
check_param SL_VLAN_PRIVATE
check_param BOSH_INIT_LOG_LEVEL

source /etc/profile.d/chruby.sh
chruby 2.2.4

semver=`cat version-semver/number`
cpi_release_name=bosh-softlayer-cpi
deployment_dir="${PWD}/deployment"
manifest_filename="director-manifest.yml"

mkdir -p $deployment_dir

cat > "${deployment_dir}/${manifest_filename}"<<EOF
---
name: bosh

releases:
- name: bosh
  url: file://bosh-release.tgz
- name: bosh-softlayer-cpi
  url: file://bosh-softlayer-cpi.tgz

resource_pools:
- name: vms
  network: default
  stemcell:
    url: file://stemcell.tgz
  cloud_properties:
    VmNamePrefix: $SL_VM_NAME_PREFIX
    Domain: $SL_VM_DOMAIN
    StartCpus: 4
    MaxMemory: 8192
    EphemeralDiskSize: 100
    Datacenter:
      Name: $SL_DATACENTER
    HourlyBillingFlag: true
    PrimaryNetworkComponent:
      NetworkVlan:
        Id: $SL_VLAN_PUBLIC
    PrimaryBackendNetworkComponent:
      NetworkVlan:
        Id: $SL_VLAN_PRIVATE

disk_pools:
- name: disks
  disk_size: 40_000

networks:
- name: default
  type: dynamic
  dns:
  - 8.8.8.8
  - 10.0.80.11
  - 10.0.80.12

jobs:
- name: bosh
  instances: 1

  templates:
  - {name: nats, release: bosh}
  - {name: postgres, release: bosh}
  - {name: blobstore, release: bosh}
  - {name: director, release: bosh}
  - {name: health_monitor, release: bosh}
  - {name: powerdns, release: bosh}
  - {name: softlayer_cpi, release: bosh-softlayer-cpi}

  resource_pool: vms
  persistent_disk_pool: disks

  networks:
  - name: default

  properties:
    nats:
      user: nats
      password: nats
      auth_timeout: 3
      address: 127.0.0.1
      listen_address: 0.0.0.0
      port: 4222
      no_epoll: false
      no_kqueue: true
      ping_interval: 5
      ping_max_outstanding: 2
      http:
        port: 9222
    postgres: &20585760
      user: postgres
      password: postgres
      host: 127.0.0.1
      database: bosh
      adapter: postgres
    blobstore:
      address: 127.0.0.1
      director:
        user: director
        password: director
      agent:
        user: agent
        password: agent
      port: 25250
      provider: dav
    director:
      cpi_job: softlayer_cpi
      address: 127.0.0.1
      name: bosh
      db:
        adapter: postgres
        database: bosh
        host: 127.0.0.1
        password: postgres
        user: postgres
      enable_virtual_delete_vms: true
    hm:
      http:
        user: hm
        password: hm
        port: 25923
      director_account:
        user: admin
        password: Cl0udyWeather
      intervals:
        log_stats: 300
        agent_timeout: 180
        rogue_agent_alert: 180
        prune_events: 30
        poll_director: 60
        poll_grace_period: 30
        analyze_agents: 60
      pagerduty_enabled: false
      resurrector_enabled: false
    dns:
      address: 127.0.0.1
      domain_name: microbosh
      db: *20585760
      webserver:
        port: 8081
        address: 0.0.0.0
    softlayer: &softlayer
      username: $SL_USERNAME
      apiKey: $SL_API_KEY

    # Tells agents how to contact nats
    agent: {mbus: "nats://nats:nats@127.0.0.1:4222"}

    ntp: &ntp []

cloud_provider:
  template: {name: softlayer_cpi, release: bosh-softlayer-cpi}

  # Tells bosh-init how to contact remote agent
  mbus: https://admin:admin@$SL_VM_NAME_PREFIX.$SL_VM_DOMAIN:6868

  properties:
    softlayer: *softlayer

    # Tells CPI how agent should listen for bosh-init requests
    agent: {mbus: "https://admin:admin@$SL_VM_NAME_PREFIX.$SL_VM_DOMAIN:6868"}

    blobstore: {provider: local, path: /var/vcap/micro_bosh/data/cache}

    ntp: *ntp
EOF

cp ./bosh-cpi-dev-artifacts/${cpi_release_name}-${semver}.tgz ${deployment_dir}/${cpi_release_name}.tgz
cp ./stemcell/*.tgz ${deployment_dir}/stemcell.tgz
cp ./bosh-release/*.tgz ${deployment_dir}/bosh-release.tgz

pushd ${deployment_dir}

  function finish {
    echo "Final state of director deployment:"
    echo "=========================================="
    cat director-manifest-state.json
    echo "=========================================="

    echo "Director:"
    echo "=========================================="
    cat /etc/hosts | grep "$SL_VM_NAME_PREFIX.$SL_VM_DOMAIN" | awk '{print $1}' | tee director-info
    echo "=========================================="

    cp -r $HOME/.bosh_init ./
  }
  trap finish ERR

  chmod +x ../bosh-init/bosh-init*
  echo "using bosh-init CLI version..."
  ../bosh-init/bosh-init* version

  echo "deploying BOSH..."
  ../bosh-init/bosh-init* deploy ${manifest_filename}

  trap - ERR
  finish
popd