#!/usr/bin/env bash

set -eu

mv director-state/.bosh "$HOME"/

state_path() { bosh-cli int director-state/director.yml --path="$1" ; }

BOSH_ENVIRONMENT=$( jq -r ".current_ip" director-state/director-state.json )
BOSH_CA_CERT=$( bosh-cli int director-state/director-creds.yml --path /director_ssl/ca )
BOSH_CLIENT_SECRET=$( bosh-cli int director-state/director-creds.yml --path /admin_password )

export BOSH_ENVIRONMENT
export BOSH_CA_CERT
export BOSH_CLIENT=admin
export BOSH_CLIENT_SECRET

set +e

bosh-cli deployments --column name | xargs -n1 -I % bosh-cli -n -d % delete-deployment
bosh-cli clean-up -n --all

if [ -f "director-state/director.yml" ] && [ -f "director-state/director-creds.yml" ]
then
	echo -e "\\n\\033[32m[INFO] Director manifest found, deleting director environment.\\033[0m"
	bosh-cli delete-env -n director-state/director.yml -l director-state/director-creds.yml
else
	echo -e "\\n\\033[33m[WARNING] Director manifest not found, You need check manually.\\033[0m"
fi
