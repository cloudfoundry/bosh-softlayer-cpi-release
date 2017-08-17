#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param SL_USERNAME
check_param SL_API_KEY

export SL_USERNAME=${SL_USERNAME}
export SL_API_KEY=${SL_API_KEY}

pushd bosh-cpi-release/src/bosh-softlayer-cpi > /dev/null
  echo "[run-lifecycle] Installing ginkgo"
  go get github.com/onsi/ginkgo/ginkgo
  echo "[run-lifecycle] Set GO15VENDOREXPERIMENT=1 for using go1.5"
  export GO15VENDOREXPERIMENT=1

  echo "[run-lifecycle] Running integration tests"
  ./bin/test-integration
popd > /dev/null
