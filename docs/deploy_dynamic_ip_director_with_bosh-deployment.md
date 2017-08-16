# Deploy dynamic IP director with bosh-deployment

## Scenario

When use BOSH CLI v2 comand `create-env` described in [doc](https://bosh.io/docs/cli-v2.html#create-env), the internal_ip is specified by option `-v internal_ip=static_ip`. This method is using static IP, and static IP before before attempting to `bosh create-env`.  
To address by more convenient method, create director enviroment by dynamic IP is following up.

## Preparation

[**Important**]Make sure to get below dependents before deploy dynamic IP director.

1. [Specified BOSH CLI](https://s3.amazonaws.com/bosh-softlayer-artifacts/bosh-cli-2.0.29-softlayer-linux-amd64): The BOSH CLI v2.0.29 enable `OSreload` function.
2. [bosh-deployment:softlayer](https://github.com/bluebosh/bosh-deployment/tree/softlayer): This repo has CPI configuration(cpi_dynamic.yml) enabling dynamic networks.

## Deployment

1. Generate director deploy manifest  
    Download BOSH CLI, chmod +x ~/Downloads/bosh-cli-* and mv to `/usr/local/bin`.  
    ```bash
    wget https://s3.amazonaws.com/bosh-softlayer-artifacts/bosh-cli-2.0.29-softlayer-linux-amd64
    chmod +x /workspace/bosh-cli-*
    mv /workspace/bosh-cli-* /usr/local/bin/bosh
    ```  
    Clone `bosh-deployment `repo to your workspace and set your env stuffs(<*>).  
    Interpolate to generate manifest.  
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
    And check `director_dynamic.yml`to ensure the `networks` section has type `type` network.
  
2. Deployment  
    Run `bosh create-env director_dynamic.yml` easily.  

3. Checking  
    Run BOSH commands to verify the deployment of director is successful.
    ```bash
    bosh -e <director_name> l
    bosh -e <director_name> vms
    ```
