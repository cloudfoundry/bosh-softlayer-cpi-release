# BOSH Softlayer CPI Release

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* BOSH Slack channel: [#bosh](https://cloudfoundry.slack.com/archives/bosh)
* BOSH SoftLayer CPI Slack channel: [#bosh-softlayer-cpi](https://cloudfoundry.slack.com/archives/bosh-softlayer-cpi)
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: <http://bosh-softlayer-cpi.ci.bluemix.net/pipelines/bosh-softlayer-cpi-release-v2>
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/1344876)

## Releases

This is a BOSH release for the Softlayer CPI.

The latest version for the SoftlLayer CPI release is here. Also, it is already available on [bosh.io](http://bosh.io).

To use this CPI you will need to use the SoftLayer light stemcell. it's also available on [bosh.io](http://bosh.io).

## Bootstrap on SoftLayer

You can use bosh-init from community to bootstrap an environment.

See [bosh-init-usage](docs/bosh-init-usage.md). Use the CPI and stemcells releases above to do so.

## Deployment Manifests Samples

For Cloud Config, see [sl-cloud-config](docs/sl-cloud-config.yml)

For Concourse, follow the guide of [Cluster with BOSH](http://concourse.ci/clusters-with-bosh.html) and reference the deployment manifest sample in ```Deploying Concourse``` section.

For a minimalistic deployment of Cloud Foundry (diego architecture), follow the introduction  [minimalistic_cf_deployment.md](docs/minimalistic_cf_deployment.md).
