# Introduction

The `minimal-softlayer.yml` is an example of a minimalistic deployment of Cloud
Foundry, including all crucial components for its basic functionality. it allows
you to deploy Cloud Foundry for educational purposes, so you can poke around and
break things. 

*IMPORTANT*: This is not meant to be used for a production level deployment as
it doesn't include features such as high availability and security.

# Setup

## Deploy BOSH director in SoftLayer
- [bosh-init-usage](docs/bosh-init-usage.md)

## Create New Subnet (Portable IPs) For Cloud Foundry Deployment
- To use the portable IPs, you need to first apply them from Softlayer: [utilizing-subnets-and-ips](https://knowledgelayer.softlayer.com/learning/utilizing-subnets-and-ips)
- Order Global IP Address: [global-ip](https://knowledgelayer.softlayer.com/topic/global-ip)
- Replace REPLACE_WITH_GLOBAL_IP in the example manifest with the new Global IP address

## DNS Configuration
If you have a domain you plan to use for your Cloud Foundry System Domain. Set up the DNS as follows:
- http://knowledgelayer.softlayer.com/articles/dns-overview-resource-records

Create a wildcard DNS entry for your root System Domain (\*.your-cf-domain.com) to point at the above Global IP address

If you do NOT have a domain, you can use 0.0.0.0.xip.io for your System Domain and replace the zeroes with your Global IP

- Replace REPLACE_WITH_SYSTEM_DOMAIN with your system domain

Generate a SSL certificate for your System Domain

- Run "openssl genrsa -out cf.key 1024"
- Run "openssl req -new -key cf.key -out cf.csr"
  - For the Common Name, you must enter "\*." followed by your System Domain
- Run "openssl x509 -req -in cf.csr -signkey cf.key -out cf.crt"
- Run "cat cf.crt && cat cf.key" and replace REPLACE_WITH_SSL_CERT_AND_KEY with this value.

## Create and Deploy Cloud Foundry
Download a copy of the latest bosh stemcell.

```sh
bosh public stemcells
bosh download public stemcell light-bosh-stemcell-[GREATEST_NUMBER]-softlayer-xen-ubuntu-trusty-go_agent.tgz

```

Upload the new stemcell.

```sh
bosh upload stemcell light-bosh-stemcell-[GREATEST_NUMBER]-softlayer-xen-ubuntu-trusty-go_agent.tgz
```

Replace REPLACE_WITH_BOSH_STEMCELL_VERSION in the example manifest with the BOSH stemcell version you downloaded above

Upload the latest stable version of Cloud Foundry.

```sh
bosh upload release https://bosh.io/d/github.com/cloudfoundry/cf-release
bosh upload release https://bosh.io/d/github.com/cloudfoundry/diego-release
bosh upload release https://bosh.io/d/github.com/cloudfoundry/cflinuxfs2-rootfs-release
bosh upload release https://bosh.io/d/github.com/cloudfoundry-incubator/garden-linux-release
bosh upload release https://bosh.io/d/github.com/cloudfoundry-incubator/etcd-release

bosh deployment LOCATION_OF_YOUR_MODIFIED_EXAMPLE_MANIFEST
bosh deploy
```

## Pushing your first application
When interacting with your Cloud Foundry deployment, you will need to download the latest stable version of the
[cf cli](https://github.com/cloudfoundry/cli). You can download a few simple applications with the following commands:

```sh
cd $HOME/workspace
git clone https://github.com/cloudfoundry/cf-acceptance-tests.git
```

For a first application, you can push a light weight ruby application called 
[Dora](https://github.com/cloudfoundry/cf-acceptance-tests/tree/master/assets/dora)
which is found at `$HOME/workspace/cf-acceptance-tests/assets/dora`. Lastly you can follow the 
[application deploy instructions](http://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html) and 
push dora into your Cloud Foundry deployment.

## SSH onto your CF Machines
Every once and a while you might want to ssh onto one of your cloud foundry deployment vms and obtain the logs
or perform some other actions. To do this you first need to get a list of vms with the command `bosh vms`. You can then
acccess your machines with the following command:

```sh
bosh ssh VM_FROM_BOSH_VMS_COMMAND/INSTANCE_NUMBER
```

Note that this command will ask you to setup the password for sudo during the login processs.

