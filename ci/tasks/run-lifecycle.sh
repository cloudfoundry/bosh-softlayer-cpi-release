#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param SL_USERNAME
check_param SL_API_KEY

export SL_USERNAME=${SL_USERNAME}
export SL_API_KEY=${SL_API_KEY}

pushd bosh-cpi-release > /dev/null
  source .envrc
  pushd src/bosh-softlayer-cpi > /dev/null
    echo -e "\n\033[32m[INFO] Installing ginkgo.\033[0m"
    go get github.com/onsi/ginkgo/ginkgo
    echo -e "\n\033[32m[INFO] Setting GO15VENDOREXPERIMENT=1 for using go1.5.\033[0m"
    export GO15VENDOREXPERIMENT=1

    echo -e "\n\033[32m[INFO] Running integration tests.\033[0m"
    ./bin/test-integration
  popd > /dev/null
popd > /dev/null
