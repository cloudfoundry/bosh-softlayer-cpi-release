#!/usr/bin/env bash

# shellcheck disable=SC1091
source /etc/profile.d/chruby.sh
chruby ruby

function cp_artifacts {
  mv "$HOME"/.bosh director-state/
  cp -r director.yml director-creds.yml director-state.json cpi-release/ director-state/
}

trap cp_artifacts EXIT

: "${BAT_INFRASTRUCTURE:?}"
: "${BOSH_SL_VM_NAME_PREFIX:?}"
: "${BOSH_SL_VM_DOMAIN:?}"

mv bosh-cli/bosh-cli-* /usr/local/bin/bosh-cli
chmod +x /usr/local/bin/bosh-cli

echo -e "\\n\\033[32m[INFO] Generating local cpi release manifest.\\033[0m"
CPI_RELEASE=$(echo cpi-release/*.tgz)
export CPI_RELEASE

cat > cpi-replace.yml <<EOF
---
- type: replace
  path: /releases/name=bosh-softlayer-cpi?
  value:
    name: bosh-softlayer-cpi
    url: file://$CPI_RELEASE

EOF

echo -e "\\n\\033[32m[INFO] Creating an empty VM for deploying BOSH director with dynamic network.\\033[0m"
./bosh-cpi-release/docs/create_vm_sl.sh -h ${BOSH_SL_VM_NAME_PREFIX} -d ${BOSH_SL_VM_DOMAIN} -c 4 -m 8192 -hb true -dc ${BOSH_SL_DATACENTER} -uv ${BOSH_SL_VLAN_PUBLIC} -iv ${BOSH_SL_VLAN_PRIVATE} -u ${BOSH_SL_USERNAME} -k ${BOSH_SL_API_KEY} > director-state.json

director_ip=$(jq -r ".current_ip" director-state.json)

echo -e "\\n\\033[32m[INFO] Generating manifest director.yml.\\033[0m"
powerdns_yml_path=$(find "./" -name powerdns.yml | head -n 1)
bosh-cli interpolate bosh-deployment/bosh.yml \
  -o "${powerdns_yml_path}" \
  -o bosh-deployment/"${BAT_INFRASTRUCTURE}"/cpi-dynamic.yml \
  -o ./cpi-replace.yml \
  -o bosh-deployment/jumpbox-user.yml \
  -o bosh-cpi-release/ci/bats/ops/remove-health-monitor.yml \
  -v internal_ip="${director_ip}" \
  -v sl_username="${BOSH_SL_USERNAME}" \
  -v sl_api_key="${BOSH_SL_API_KEY}" \
  -v sl_datacenter="${BOSH_SL_DATACENTER}" \
  -v sl_vlan_private="${BOSH_SL_VLAN_PRIVATE}" \
  -v sl_vlan_public="${BOSH_SL_VLAN_PUBLIC}" \
  -v sl_vm_name_prefix="${BOSH_SL_VM_NAME_PREFIX}" \
  -v sl_vm_domain="${BOSH_SL_VM_DOMAIN}" \
  -v dns_recursor_ip=8.8.8.8 \
  -v director_name=bats-director \
  --vars-store director-creds.yml \
  > director.yml

export BOSH_LOG_LEVEL=DEBUG
export BOSH_LOG_PATH=./run.log
echo -e "\\n\\033[32m[INFO] Deploying director.\\033[0m"
bosh-cli create-env \
  --state director-state.json \
  --vars-store director-creds.yml \
  director.yml

cat > cloud_config.yml <<EOF
---
azs:
- cloud_properties:
    datacenter: lon02
  cpi: softlayer
  name: z1
compilation:
  az: z1
  network: default
  reuse_compilation_vms: true
  vm_type: default
  workers: 2
networks:
- name: default
  subnets:
  - az: z1
    cloud_properties:
      security_groups:
      - default
      - cf
      vlan_ids:
      - 1292653
      - 1292651
    dns:
    - 10.0.80.11
    - 10.0.80.12
  type: dynamic
vm_types:
- cloud_properties:
    cpu: 2
    ephemeral_disk_size: 100
    hostname_prefix: slcpi-default-
    hourly_billing_flag: true
    local_disk_flag: true
    memory: 4096
  name: default

EOF

cat > cpi_config.yml <<EOF
---
cpis:
- name: softlayer
  type: softlayer

EOF

export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET=$(bosh-cli int director-creds.yml --path /admin_password)
bosh-cli -e ${director_ip} --ca-cert <(bosh-cli int director-creds.yml --path /director_ssl/ca) alias-env ${director_ip}
bosh-cli -e ${director_ip} -n update-cloud-config cloud_config.yml
bosh-cli -e ${director_ip} -n update-cpi-config cpi_config.yml
