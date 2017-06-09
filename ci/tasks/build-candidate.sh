#!/usr/bin/env bash

set -e

semver=`cat version-semver/number`

BOSH_CLI="$(pwd)/$(echo bosh-cli/bosh-cli-*)"
chmod +x ${BOSH_CLI}

pushd bosh-cpi-release

  source .envrc
  pushd src/bosh-softlayer-cpi > /dev/null
    echo "[build-candidate] Installing ginkgo"
    go get github.com/onsi/ginkgo/ginkgo
    echo "[build-candidate] Set GO15VENDOREXPERIMENT=1 for using go1.5"
    export GO15VENDOREXPERIMENT=1

    echo "[build-candidate] Running unit tests"
    ./bin/test-unit
  popd > /dev/null

  echo $semver > src/bosh-softlayer-cpi/version

  cpi_release_name="bosh-softlayer-cpi"
  tarball_name="dev_releases/${cpi_release_name}/${cpi_release_name}-${semver}.tgz"
  echo "[build-candidate] Building CPI release..."

  $BOSH_CLI create-release --name $cpi_release_name --version $semver --tarball $tarball_name --force
popd

mv bosh-cpi-release/$tarball_name candidate/
