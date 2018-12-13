# BOSH Director

##1. Update bosh director yml template
1. Convert cloud properties names to align with CPI ng. Refer to [cloud_properties](cloud_properties_names_in_cpi_ng.md)
2. Use new CPI release - TBD
3. Use new stemcell version - TBD
4. Add registry job
  - name: registry
    release: bosh

## 2. Upgrade bosh director

The cloud_properties names in CPI ng changed. Need to convert cloud-config (azs, networks and vm_types sections) to align with CPI ng.
All Public environments have cloud config ready. If the target env does't have cloud config, make it ready first.
Download the [script](https://github.com/bluebosh/bosh-softlayer-tools/blob/master/scripts/convert_cloudconfig.sh) to convert cloud-config to align with CPI ng

1. Download cloud-config from bosh director
on boshcli,
bosh2 cloud-config >  cloud-config.yml

2. Convert the  cloud-config YAML file
./convert_cloudconfig.sh cloud-config.yml
