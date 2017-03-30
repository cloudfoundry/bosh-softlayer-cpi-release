# Recover the missing vm from 'bosh deploy' failure


# Table of Contents

- [Purpose](#purpose)
- [User Impact](#user_impact)
- [Instructions to Fix](#instructions)
  - [Scenario 1: user_data.json exists](#scenario1)
  - [Scenario 2: user_data.json does not exist](#scenario2)

# <a id="purpose"></a>Purpose
When `bosh deploy` fails, sometimes the failing vm is missing from `bosh vms` (deleted from bosh db) but still alive in Softlayer. If you do nothing, in the next `bosh deploy`, bosh will treat it as a missing vm and create a new one with different IP. Sometimes we need to keep the vm IP unchanged, so the missing vm needs to be recovered in the bosh db.

# <a id="user_impact"></a>User Impact
If you don't perform this run book, the next `bosh deploy` will create a new vm to backfill the missing vm with different IP. This will break the cases when the vm IP needs to be kept unchanged. 

# <a id="instructions"></a>Instructions to Fix

- ## <a id="scenario1"></a>Scenario 1: user_data.json exists

    Make sure the failing vm can still been seen from [Softlayer Portal](https://control.softlayer.com) and `/var/vcap/bosh/user_data.json` exists on the failing vm.

    Take `nats/0` as an example of the failing vm.

    ### 1) Update the vm info in "instances" table of bosh db

    ```
    update instances set vm_cid='12345678' where job='nats' and index=0;
    update instances set agent_id='1234abcd-12ab-34cd-56ef-123456abcdef' where job='nats' and index=0;
    ```
    The `vm_cid` can be got from Softlayer Portal address such as `12345678` in https://control.softlayer.com/devices/details/12345678/virtualGuest
    
    The `agent_id` can be got from `/var/vcap/bosh/user_data.json` on the failing vm.
     
    ### 2) Create spec.json for the failing vm
    
    Login to the failing vm and create /var/vcap/bosh/spec.json like this:
    
    ```
    {
      "job": {
        "name": "REPLACE_job_name"  
      },
      "deployment": "REPLACE_deployment_name",
      "networks": {
        "default": {
          "cloud_properties": {
            "security_groups": [
              "default",
              "cf"
            ]
          },
          "default": [
            "dns",
            "gateway"
          ],
          "dns": [
            "REPLACE_dns_ip",
            "10.0.80.11",
            "10.0.80.12"
          ],
          "gateway": "REPLACE_gateway_ip",
          "ip": "REPLACE_vm_ip",
          "netmask": "REPLACE_netmask_ip",
          "type": "dynamic"
        }
      },
      "index": REPLACE_job_index,
      "id": "",
      "persistent_disk": 0
    } 
    ```
    
    - **name**: Replace with the job name. It's `nats` in the example.
    - **dns**: Replace with the dns ip. In case the PowerDNS on director is used, it's the director ip.
    - **gateway**: Replace with the gateway ip. It can be found from `route -n` or `netstat -rn`. See example below.
    - **ip**: Replace with the vm ip. It can be found from `ifconfig`. See example below.
    - **netmask**: Replace with the netmask ip. It can be found from `ifconfig`. See example below.
    - **deployment**: Replace with the name of deployment which the vm belongs to.
    - **index**: Replace with the job index. It's `0` in the example.
    
  *`netstat -rn` or `route -n` output example:*
    ```
    Kernel IP routing table
    Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
    0.0.0.0         <gateway_ip>    0.0.0.0         UG    0      0        0 eth1
    9.0.0.0         10.121.120.193  255.0.0.0       UG    0      0        0 eth0
    10.0.0.0        10.121.120.193  255.0.0.0       UG    0      0        0 eth0
    10.121.120.192  0.0.0.0         255.255.255.192 U     0      0        0 eth0
    161.26.0.0      10.121.120.193  255.255.0.0     UG    0      0        0 eth0
    169.53.1.128    0.0.0.0         255.255.255.224 U     0      0        0 eth1
    ```
    
  *`ifconfig` output example:*
    ```
    eth0      Link encap:Ethernet  HWaddr 06:99:32:27:34:f7
              inet addr:<vm_private_ip>  Bcast:10.121.120.255  Mask:<netmask_ip_private>
              inet6 addr: fe80::499:32ff:fe27:34f7/64 Scope:Link
              UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
              RX packets:596214 errors:0 dropped:0 overruns:0 frame:0
              TX packets:5065 errors:0 dropped:0 overruns:0 carrier:0
              collisions:0 txqueuelen:1000
              RX bytes:63759564 (63.7 MB)  TX bytes:661260 (661.2 KB)
     
    eth1      Link encap:Ethernet  HWaddr 06:da:3f:89:2a:b2
              inet addr:<vm_public_ip>  Bcast:169.53.1.159  Mask:<netmask_ip_public>
              inet6 addr: fe80::4da:3fff:fe89:2ab2/64 Scope:Link
              UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
              RX packets:65584 errors:0 dropped:0 overruns:0 frame:0
              TX packets:21240 errors:0 dropped:0 overruns:0 carrier:0
              collisions:0 txqueuelen:1000
              RX bytes:5413249 (5.4 MB)  TX bytes:3236947 (3.2 MB)
     
    lo        Link encap:Local Loopback
              inet addr:127.0.0.1  Mask:255.0.0.0
              inet6 addr: ::1/128 Scope:Host
              UP LOOPBACK RUNNING  MTU:65536  Metric:1
              RX packets:6291 errors:0 dropped:0 overruns:0 frame:0
              TX packets:6291 errors:0 dropped:0 overruns:0 carrier:0
              collisions:0 txqueuelen:0
              RX bytes:892748 (892.7 KB)  TX bytes:892748 (892.7 KB)
    ``` 
    
    ### 3) Restart bosh agent on the failing vm
    `sv restart agent`
     
    ### 4) Make sure no problem with "bosh vms" and "bosh cck"
    - `bosh vms`: the missing vm is back
    - `bosh cck`: no problems
     
    ### 5) Re-run "bosh deploy"

---
- ## <a id="scenario2"></a>Scenario 2: user_data.json does not exist
    Sometimes when `bosh deploy` fails, user_data.json has not yet been generated in cases like create_vm failure.
    
    Make sure the failing vm can still been seen from [Softlayer Portal](https://control.softlayer.com)

    ### 1) Do the steps 1) & 2) in [Scenario 1](#scenario1)
    Except the `agent_id` needs need to be got from the bosh debug log for the failure.
    
    **bosh debug log example:**
    
    *E, [2016-07-22 06:17:15 #13984] [canary_update(nats/0 (14109da6-cc51-49e8-8e99-d096e4f59550))] ERROR -- DirectorJobRunner: error creating vm: Creating Virtual_Guest with agent ID '**1234abcd-12ab-34cd-56ef-123456abcdef**': Attaching ephemeral disk to VirtualGuest \`12345678\`: Waiting for VirtualGuest \`12345678\` has Service Setup transaction complete: Getting Last Complete Transaction for virtual guest with ID '12345678': Get https://[user]:[api_key]@api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest/12345678/getLastTransaction.json?objectMask=transactionGroup: net/http: TLS handshake timeout*
 
    ### 2) Create /var/vcap/bosh/user_data.json
    
    **`user_data.json` example:**
    
    *{"agent_id":"**1234abcd-12ab-34cd-56ef-123456abcdef**","vm":{"name":"vm-**1234abcd-12ab-34cd-56ef-123456abcdef**","id":"vm-**1234abcd-12ab-34cd-56ef-123456abcdef**"},"mbus":"nats://nats:nats@**[director_ip]**:4222","ntp":[],"blobstore":{"provider":"dav","options":{"endpoint":"http://**[director_ip]**:25250","password":"[agent_password]","user":"agent"}},"networks":{"default":{"type":"dynamic","ip":"**[vm_ip]**","netmask":"**[netmask]**","gateway":"**[gateway_ip]**","dns":["**[dns_ip]**","10.0.80.11","10.0.80.12"],"default":["dns","gateway"],"preconfigured":true,"cloud_properties":{"security_groups":["default","cf"]}}},"disks":{"ephemeral":"/dev/xvdc","persistent":{}},"env":{}}*
    
    Update the items in **bold**: 
    - `agent_id` can be found in the bosh debug log
    - Keep "persistent" as empty which will be fixed later)
    - Refer to [Scenario 1](#scenario1) to get `vm_ip`, `netmask` and `gateway_ip`
 
    ### 3) Restart bosh agent
    `sv restart agent`
 
    ### 4) Run `bosh cck` and fix the persistent disk problem if exists
    ```
    Problem 1 of 1: Inconsistent mount information:
    Record shows that disk '12661853' should be mounted on 22354779.   
    However it is currently :    
        Not mounted in any VM.    
      1. Ignore    
      2. Reattach disk to instance    
      3. Reattach disk and reboot instance    
    Please choose a resolution [1 - 3]: 2
    ```
 
    ### 5) Make sure no problem with "bosh vms" and "bosh cck"
    - `bosh vms`: the missing vm is back
    - `bosh cck`: no problems
 
    ### 6) Re-run `bosh deploy`


