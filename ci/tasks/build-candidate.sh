#!/usr/bin/env bash

set -e

semver=`cat version-semver/number`

mv bosh-cli/bosh-cli-* /usr/local/bin/bosh-cli
chmod +x /usr/local/bin/bosh-cli

pushd bosh-cpi-release
  source .envrc
  pushd src/bosh-softlayer-cpi > /dev/null
    echo -e "\n\033[32m[INFO] Installing ginkgo.\033[0m"
    go get github.com/onsi/ginkgo/ginkgo
    echo -e "\n\033[32m[INFO] Set GO15VENDOREXPERIMENT=1 for using go1.5.\033[0m"
    export GO15VENDOREXPERIMENT=1

    echo -e "\n\033[32m[INFO] Running unit tests and updating coverage.\033[0m"
    ./bin/test-unit-cover
  popd > /dev/null

  echo $semver > src/bosh-softlayer-cpi/version

  cpi_release_name="bosh-softlayer-cpi"
  tarball_name="dev_releases/${cpi_release_name}/${cpi_release_name}-${semver}.tgz"

  echo -e "\n\033[32m[INFO] Building CPI release.\033[0m"
  bosh-cli create-release --name $cpi_release_name --version $semver --tarball $tarball_name --force
popd

checksum="$(sha1sum "./bosh-cpi-release/$tarball_name" | awk '{print $1}')"
echo "$tarball_name sha1=$checksum"

mv bosh-cpi-release/$tarball_name candidate/
