# Table of Contents

1. [1. 'bosh deploy' fails and the failing vm is missing from 'bosh vms', but you want to keep it (user_data.json exists)](#p1)
2. [Problem2](#p2)
3. [Problem3](#p3)

---

This document shows the tips when run `bosh deploy` on Softlayer.

## <a id="p1"></a>1. 'bosh deploy' fails and the failing vm is missing from 'bosh vms', but you want to keep it (user_data.json exists) ####

Make sure the failing vm can still been seen from [Softlayer Portal](https://control.softlayer.com) and `/var/vcap/bosh/user_data.json` exists on the failing vm.

Take `nats/0` as an example of the failing vm.

### 1) Update the vm info in "instances" table of bosh db

```
update instances set vm_cid='12345678' where job='nats' and index=0;
update instances set agent_id='1234abcd-12ab-34cd-56ef-123456abcdef' where job='nats' and index=0;
```
The `vm_cid` can be got from Softlayer Portal address such as https://control.softlayer.com/devices/details/12345678/virtualGuest

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

**`netstat -rn` or `route -n` output example:**
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

**`ifconfig` output example:**
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



## <a id="p2"></a>2. Problem2


## <a id="p3"></a>3. Problem3
