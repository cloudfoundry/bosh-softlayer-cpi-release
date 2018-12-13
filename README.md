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

## Migrate from Legacy SoftLayer CPI to SoftLayer CPI NG

For the users using legacy SoftLayer CPI who want to move to SoftLayer CPI NG, please refer to [softlayer-cpi-migration](docs/softlayer-cpi-migration.md) to do the migration.

## Frequently Asked Questions and Answers

1. Q: How do I specify a dynamic network through subnet instead of vlan id?

   A: We don't support it currently.

2. Q: Is there any restrictions about the hostname supported by SoftLayer?

   A: Yes. The hostname length can't be exactly 64. Otherwise there will be problems with ssh login.
