## Experimental `bosh-init` usage on SoftLayer

At present, it is able to create and update bosh environment on SoftLayer by bosh-init and bosh-softlayer-cpi

This document shows how to initialize new environment on SoftLayer.

0. Create a deployment directory, install bosh-init

http://bosh.io/docs/install-bosh-init.html

```
mkdir my-bosh-init
```

1. Create a deployment manifest file named bosh.yml in the deployment directory based on the template below.

```
---
name: bosh

releases:
- name: bosh
  url: https://bosh.io/d/github.com/cloudfoundry/bosh?v=256.2
  sha1: ff2f4e16e02f66b31c595196052a809100cfd5a8
- name: bosh-softlayer-cpi
  url: https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-softlayer-cpi-release?v=2.3.0
  sha1: ed79e5617961076a22c8111d173bc5bfeb6a9b14
  
resource_pools:
- name: vms
  network: default
  stemcell:
    url: file://./light-bosh-stemcell-3232.4-softlayer-esxi-ubuntu-trusty-go_agent.tgz 
  cloud_properties:
    Domain: softlayer.com
    VmNamePrefix: bosh-softlayer
    EphemeralDiskSize: 100
    StartCpus: 4
    MaxMemory: 8192
    Datacenter:
       Name: lon02
    HourlyBillingFlag: true
    PrimaryNetworkComponent:
       NetworkVlan:
          Id: <pub vlan>
    PrimaryBackendNetworkComponent:
       NetworkVlan:
          Id: <pri vlan>
    NetworkComponents:
    - MaxSpeed: 1000
disk_pools:
- name: disks
  disk_size: 40_000

networks:
- name: default
  type: dynamic
  dns: 
  - 8.8.8.8

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
      password: postges
      host: 127.0.0.1
      database: bosh
      adapter: postgres
    blobstore:
      address: 127.0.0.1
      director:
        user: director
        password: <password>
      agent:
        user: agent
        password: <password>
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
        password: postges
        user: postgres
    hm:
      http:
        user: hm
        password: hm
        port: 25923
      director_account:
        user: admin
        password: <password>
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
      username: <sl_key_user> # <--- Replace with username
      apiKey: <sl_api_key> # <--- Replace with password

    agent: {mbus: "nats://nats:nats@127.0.0.1:4222"}

    ntp: &ntp []
 
cloud_provider:
  template: {name: softlayer_cpi, release: bosh-softlayer-cpi}

  # Tells bosh-init how to contact remote agent
  mbus: https://admin:admin@bosh-softlayer.softlayer.com:6868

  properties:
    softlayer: *softlayer

    # Tells CPI how agent should listen for bosh-init requests
    agent: {mbus: "https://admin:admin@bosh-softlayer.softlayer.com:6868"}

    blobstore: {provider: local, path: /var/vcap/micro_bosh/data/cache}

    ntp: *ntp

```

2. Kick off a deployment

```
bosh-init deploy bosh.yml
```

**Notes:**

**1. Please make sure to run bosh-init with root user, since there needs to update /etc/hosts.**

**2. Please note that there needs communication between bosh-init and the target director VM over SoftLayer private network to accomplish a successful deployment. So please make sure the machine where to run bosh-init can access SoftLayer private network. You could enable [SoftLayer VPN](http://www.softlayer.com/VPN-Access) if the machine where to run bosh-init is outside SoftLayer data center.**
