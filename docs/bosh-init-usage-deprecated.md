## Experimental `bosh-init` usage on SoftLayer

At present, it is able to create and update bosh environment on SoftLayer by bosh-init and bosh-softlayer-cpi

This document shows how to initialize new environment on SoftLayer.

0. Create a deployment directory, install bosh-init

http://bosh.io/docs/install-bosh-init.html

```
mkdir my-bosh-init
```

1. Create a deployment manifest file named sl-bosh.yml in the deployment directory based on [sl-bosh.yml](sl-bosh.yml)

2. Kick off a deployment

```
bosh-init deploy sl-bosh.yml
```

**Notes:**

**1. Please make sure to run bosh-init with root user, since there needs to update /etc/hosts.**

**2. Please note that there needs communication between bosh-init and the target director VM over SoftLayer private network to accomplish a successful deployment. So please make sure the machine where to run bosh-init can access SoftLayer private network. You could enable [SoftLayer VPN](http://www.softlayer.com/VPN-Access) if the machine where to run bosh-init is outside SoftLayer data center.**
