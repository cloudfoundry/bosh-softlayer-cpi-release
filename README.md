# BOSH Softlayer CPI Release

| Concourse Job      | Sub Job | Status                                                                                                                                                                                                                               |
| ---                | ---     | ---                                                                                                                                                                                                                             |
| BATS               | bats-ubuntu |   [![bosh-sl-ci.bluemix.net](https://bosh-sl-ci.bluemix.net/api/v1/teams/main/pipelines/wip:lite:bosh:softlayer:cpi/jobs/bats-ubuntu/badge)](https://bosh-sl-ci.bluemix.net/teams/main/pipelines/wip:lite:bosh:softlayer:cpi/jobs/bats-ubuntu) |
| integration        | run-lifecycle | [![bosh-sl-ci.bluemix.net](https://bosh-sl-ci.bluemix.net/api/v1/teams/main/pipelines/wip:lite:bosh:softlayer:cpi/jobs/run-lifecycle/badge)](https://bosh-sl-ci.bluemix.net/teams/main/pipelines/wip:lite:bosh:softlayer:cpi/jobs/run-lifecycle) |
| unit tests         | build-candidate | [![bosh-sl-ci.bluemix.net](https://bosh-sl-ci.bluemix.net/api/v1/teams/main/pipelines/wip:lite:bosh:softlayer:cpi/jobs/build-candidate/badge)](https://bosh-sl-ci.bluemix.net/teams/main/pipelines/wip:lite:bosh:softlayer:cpi/jobs/build-candidate) |

* Documentation: [bosh.io/docs](https://bosh.io/docs)
* BOSH SoftLayer CPI Slack channel: #bosh-softlayer-cpi on [https://cloudfoundry.slack.com/](https://cloudfoundry.slack.com/)
* Mailing list: [cf-bosh](https://lists.cloudfoundry.org/pipermail/cf-bosh)
* CI: <https://bosh-sl-ci.bluemix.net/teams/main/pipelines/wip:lite:bosh:softlayer:cpi>
* Roadmap: [Pivotal Tracker](https://www.pivotaltracker.com/n/projects/1344876)

This is a BOSH release for the Softlayer CPI.

See [Initializing a BOSH environment on Softlayer](https://bosh.io/docs/init-softlayer.html) for example usage.

## Development

See [development doc](docs/README.md).
