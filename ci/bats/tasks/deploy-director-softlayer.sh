#!/usr/bin/env bash

source /etc/profile.d/chruby.sh
chruby 2.1.7

function cp_artifacts {
  mv $HOME/.bosh director-state/
  cp -r director.yml director-creds.yml director-state.json cpi-release/ director-state/
}

trap cp_artifacts EXIT

: ${BAT_INFRASTRUCTURE:?}
: ${BOSH_SL_VM_NAME_PREFIX:?}
: ${BOSH_SL_VM_DOMAIN:?}

mv bosh-cli/bosh-cli-* /usr/local/bin/bosh-cli
chmod +x /usr/local/bin/bosh-cli

echo -e "\n\033[32m[INFO] Generating local cpi release manifest.\033[0m"
export CPI_RELEASE=$(echo cpi-release/*.tgz)

cat > cpi-replace.yml <<EOF
---
- type: replace
  path: /releases/name=bosh-softlayer-cpi?
  value:
    name: bosh-softlayer-cpi
    url: file://$CPI_RELEASE

EOF

echo -e "\n\033[32m[INFO] Generating manifest director.yml.\033[0m"
powerdns_yml_path=$(find ${pwd} -name powerdns.yml | head -n 1)
bosh-cli interpolate bosh-deployment/bosh.yml \
  -o bosh-deployment/$BAT_INFRASTRUCTURE/cpi.yml \
  -o ./cpi-replace.yml \
  -o ${powerdns_yml_path} \
  -o bosh-deployment/jumpbox-user.yml \
  -o bosh-cpi-release/ci/bats/ops/remove-health-monitor.yml \
  -v dns_recursor_ip=8.8.8.8 \
  -v director_name=bats-director \
  -v sl_director_fqn=$BOSH_SL_VM_NAME_PREFIX.$BOSH_SL_VM_DOMAIN \
  --vars-file <( bosh-cpi-release/ci/bats/iaas/$BAT_INFRASTRUCTURE/director-vars ) \
  --vars-store director-creds.yml \
  > director.yml

export BOSH_LOG_LEVEL=DEBUG
export BOSH_LOG_PATH=./run.log
echo -e "\n\033[32m[INFO] Deploying director.\033[0m"
bosh-cli create-env \
  --state director-state.json \
  --vars-store director-creds.yml \
  director.yml
