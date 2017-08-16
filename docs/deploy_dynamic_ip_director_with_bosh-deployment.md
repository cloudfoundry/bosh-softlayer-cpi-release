# Deploy dynamic IP director with bosh-deployment

## Requirement

The general way provided in [bosh-deployment](https://github.com/cloudfoundry/bosh-deployment) project is to deploy a BOSH director with static IP, but there still has strong needs to deploy a BOSH director with dynamic IP on Softlayer. This instruction will show you how to deploy a dynamic IP BOSH director with [bosh-deployment](https://github.com/bluebosh/bosh-deployment/tree/softlayer) on Softlayer. 

## Preparation

[**Important**]Make sure to use `BOSH CLI for Softlayer` and `bosh-deployment for Softlayer` repo to deploy a dynamic IP director.

1. [BOSH CLI for Softlayer](https://s3.amazonaws.com/bosh-softlayer-artifacts/bosh-cli-2.0.29-softlayer-linux-amd64): The BOSH CLI v2.0.29 with `OS Reload` function enabled.
2. [bosh-deployment:softlayer](https://github.com/bluebosh/bosh-deployment/tree/softlayer): This repo has CPI configuration(cpi_dynamic.yml) enabling dynamic network.

## Deployment

1. Generate BOSH director deployment manifest  
    
    Download `BOSH CLI for Softlayer`, and grant execution permission, better to put it in PATH:
    
    ```bash
    wget https://s3.amazonaws.com/bosh-softlayer-artifacts/bosh-cli-2.0.29-softlayer-linux-amd64
    chmod +x /workspace/bosh-cli-*
    mv /workspace/bosh-cli-* /usr/local/bin/bosh
    ```  
    Clone [bosh-deployment for Softlayer](https://github.com/bluebosh/bosh-deployment) into your workspace and and switch to `softlayer` branch.
    
    Generate deployment manifest with `bosh interpolate` command:
    ```bash
    bosh int /workspace/bosh-deployment/bosh.yml \
        -o /workspace/bosh-deployment/softlayer/cpi_dynamic.yml \
        --vars-store credentials.yml \
        -v sl_vm_name_prefix=<director_hostname> \
        -v sl_vm_domain=<director_domain>  \
        -v sl_username=<sl_username> \
        -v sl_api_key=<sl_api_lkey> \
        -v sl_datacenter=<datacenter> \
        -v sl_vlan_public=<sl_vlan_public> \
        -v sl_vlan_private=<sl_vlan_private> \
        -v director_name=<director_name> \
        -v internal_ip=<director_hostname>.<director_domain> > director_dynamic.yml
    ```
  
2. Deployment  
    Run `bosh create-env director_dynamic.yml` easily.  

3. Login and verify the BOSH environment

    ```bash
    bosh alias-env <alias_name> -e <director_hostname>.<director_domain> --ca-cert <(bosh int credentials.yml --path /director_ssl/ca)`
    export BOSH_CLIENT=admin
    export BOSH_CLIENT_SECRET=`bosh int credentials.yml --path /admin_password`

    bosh -e <alias_name> login
    
    bosh -e <director_name> vms
    ```
    
