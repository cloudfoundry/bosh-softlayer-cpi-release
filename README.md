# BOSH SoftLayer CPI Release

Coverage status: [![Coverage Status](https://coveralls.io/repos/github/cloudfoundry/bosh-softlayer-cpi-release/badge.svg?branch=master)](https://coveralls.io/github/cloudfoundry/bosh-softlayer-cpi-release?branch=master)

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* BOSH SoftLayer CPI Slack channel: #bosh-softlayer-cpi on [https://cloudfoundry.slack.com/](https://cloudfoundry.slack.com/)
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: <https://bosh-sl-ci.bluemix.net/teams/main/pipelines/bosh:softlayer:cpi_ng>
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/1344876)

## Releases

This is a BOSH release for the SoftLayer CPI.

The latest version for the SoftlLayer CPI release is available on [bosh.io](https://bosh.io/releases/github.com/cloudfoundry/bosh-softlayer-cpi-release?all=1).

To use this CPI you will need to use the SoftLayer light stemcell. it's also available on [bosh.io](https://bosh.io/stemcells/bosh-softlayer-xen-ubuntu-xenial-go_agent).

## Bootstrap on SoftLayer

Refer to [init-softlayer](docs/init-softlayer.md) to bootstrap on Softlayer.

## Deployment Manifest Samples

Refer to [softlayer-cpi](docs/softlayer-cpi.md) for deployment manifest samples.

## SoftLayer CPI NG

* Highlights of SoftLayer CPI NG: [highlight_of_cpi_ng](docs/highlight_of_cpi_ng.md)
* Cloud properties names in SoftLayer CPI NG: [cloud_properties_names_in_cpi_ng](docs/cloud_properties_names_in_cpi_ng.md)
* Migrate from Legacy SoftLayer CPI to SoftLayer CPI NG: [softlayer-cpi-migration](docs/softlayer-cpi-migration.md)
* How to use static IPs: [static-ip-example](docs/static-ip-example.md)

## Recover VMs or Disks (for legacy SoftLayer CPI only)

* Recover missing VMs: [recover_missing_vm](docs/recover_missing_vm.md)
* Recover orphan disks: [recover_orphan_disk_for_legacy_cpi](docs/recover_orphan_disk_for_legacy_cpi.md)

## Frequently Asked Questions and Answers

1. Q: How do I specify a dynamic network through subnet instead of vlan id?

   A: We don't support it currently.

2. Q: Is there any restrictions about the hostname supported by SoftLayer?

   A: Yes. The hostname length can't be exactly 64. Otherwise there will be problems with ssh login.
