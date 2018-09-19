#!/bin/bash
set -e

# Usage statement
usage() {
    echo "Create a virtual machine on IBM Cloud Infrastructure"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo
    echo "OPTIONS: "
    echo "   -h  --hostname <hostname of VM>"
    echo "                The hostname of virtual machine to be created, required"        
    echo "   -d  --domain <domain namet>"
    echo "                The domain name of virtual machine to be created, default value: softlayer.com"
    echo "   -c  --cpu <cpu core numaber>"
    echo "                The core number of CPU, default value: 2"
    echo "   -m  --memory <memory size in MB>"
    echo "                The memory size in mega byte, default value: 4096"
    echo "   -hb  --hourlyBilling <true|false>"
    echo "                Sets hourly billing if ture, montly billing if false, defalut value: true"
    echo "   -ld  --localDisk <true|false>"
    echo "                Sets to use local Disk if true, SAN disk if false, defalut value: true"
    echo "   -da --dedicatedAccountHostOnly <true|false>"
    echo "                Sets to use a dedicated host or a shared host to run the VM, default value: false"
    echo "   -dc  --datacenter <datacenter>"
    echo "                Sets the datacneter, required"
    echo "   -ms  --maxSpeed <10|100|1000>"
    echo "                Sets the network speed to 10, 100 or 1000, default value: 100"
    echo "   -uv  --publicVlan <public vlan id>"
    echo "                The public vlan id, required"
    echo "   -iv  --privateVlan <private vlan id>"
    echo "                The private vlan id, required"
    echo "   -u  --username <user name of API key>"
    echo "                The user name used to create VM on IBM Cloud Infrastructure, required"
    echo "   -k  --apiKey <API key>"
    echo "                The API key used to create VM on IBM Cloud Infrastructure, required"
    echo "   -j  --json <output in json format>"
    echo "                Return private ip in json format or a plain string, default is a plain string"
    echo
    echo "  --help  Output this help."
    echo
}


if [ $# -eq 0 ]; then
    usage
    exit 2
fi

while [[ $# -gt 0 ]]
do
    key="$1"
    shift
    case $key in
        -h|--hostname)
        hostname=$1
        shift
        ;;
        -d|--domain)
        domain=$1
        shift
        ;;
        -c|--cpu)
        cpu=$1
        shift
        ;;
        -m|--memory)
        memory=$1
        shift
        ;;
        -hb|--hourlyBilling)
        hourlyBilling=$1
        shift
        ;;
        -ld|--localDisk)
        localDisk=$1
        shift
        ;;
        -da|--dedicatedAccountHostOnly)
        dedicatedAccountHostOnly=$1
        shift
        ;;
        -dc|--datacenter)
        datacenter=$1
        shift
        ;;
        -ms|--maxSpeed)
        maxSpeed=$1
        shift
        ;;
        -uv|--publicVlan)
        publicVlan=$1
        shift
        ;;
        -iv|--privateVlan)
        privateVlan=$1
        shift
        ;;
        -u|--username)
        username=$1
        shift
        ;;
        -k|--apiKey)
        apiKey=$1
        shift
        ;; 
        -j|--json)
        json=1
        ;;                    
        *)
        echo "Unrecognized parameter(s): $*"
        usage
        exit 2
        ;;
    esac
done

# Check parameters and set default values
if [ -z "${hostname:-}" ]; then
    hostname='create-a-new-vm-test'
fi
if [ -z "${domain:-}" ]; then
    domain='softlayer.com'
fi
if [ -z "${cpu:-}" ]; then
    cpu=2
fi
if [ -z "${memory:-}" ]; then
    memory=4096
fi
if [ -z "${hourlyBilling:-}" ]; then
    hourlyBilling=true
fi
if [ -z "${localDisk:-}" ]; then
    localDisk=true
fi
if [ -z "${dedicatedAccountHostOnlyFlag:-}" ]; then
    dedicatedAccountHostOnly=false
fi
if [ -z "${maxSpeed:-}" ]; then
    maxSpeed=100
fi
if [ -z "${publicVlan:-}" ]; then
    echo "[ERROR] Empty public VLAN, please set \"-uv <vlan id>\""
    exit 1
fi
if [ -z "${privateVlan:-}" ]; then
    echo "[ERROR] Empty private VLAN, please set \"-iv <vlan id>\""
    exit 1
fi
if [ -z "${username:-}" ]; then
    echo "[ERROR] Empty username, please set \"-u <username>\""
    exit 1
fi
if [ -z "${apiKey:-}" ]; then
    echo "[ERROR] Empty api_key, please set \"-k <api key>\""
    exit 1
fi
if [ -z "${datacenter:-}" ]; then
    echo "[ERROR] Empty dataceneter, please set \"-dc <datacenter>\""
    exit 1
fi
if [ -z "${json:-}" ]; then
    json=0
fi
if [[ "${username}" =~ "@" ]]; then 
    username=${username/@/%40}
fi

vm_id_json=$(curl -X POST -d "{
  \"parameters\":
  [
    {
      \"complexType\":\"SoftLayer_Virtual_Guest\",
      \"hostname\":\"${hostname}\",
      \"domain\":\"${domain}\",
      \"startCpus\":\"${cpu}\",
      \"maxMemory\":\"${memory}\",
      \"hourlyBillingFlag\":\"${hourlyBilling}\",
      \"localDiskFlag\":\"${localDisk}\",
      \"dedicatedAccountHostOnlyFlag\":\"${dedicatedAccountHostOnly}\",
      \"blockDeviceTemplateGroup\": {
        \"globalIdentifier\":\"9975997f-ac5a-49e7-8409-fb260261dcd1\"
      },
      \"datacenter\":{
        \"name\":\"${datacenter}\"
      },
      \"networkComponents\":[
        {
          \"maxSpeed\":\"${maxSpeed}\"
        }
      ],
      \"privateNetworkOnlyFlag\":false,
      \"primaryBackendNetworkComponent\":{
        \"networkVlan\":{
          \"id\":\"${privateVlan}\"
        }
      },
      \"primaryNetworkComponent\":{
        \"networkVlan\":{
          \"id\":\"${publicVlan}\"
        }      
      }
    }
  ]
}" https://"${username}":"${apiKey}"@api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest/createObject?objectMask=id )

vm_id=$(echo "${vm_id_json}" | cut -d ":" -f2 | cut -d "}" -f1)

sleep 120

START_TIMESTAMP=$(date +%s)
declare -r START_TIMESTAMP
declare -r timeout=1800

while true
do
    trans=$(curl https://"${username}":"${apiKey}"@api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest/"${vm_id}"/getActiveTransactions)
    if [ "${trans}" != "[]" ]; then
        if [[ $(($(date +%s) - START_TIMESTAMP)) -ge ${timeout} ]]; then
            echo "Timeout ... (elapsed time $(($(date +%s) - START_TIMESTAMP)) s)"
            exit 1
        else
            sleep 15
        fi
    else
        break
    fi
done

vm_private_ip_json=$(curl https://"${username}":"${apiKey}"@api.softlayer.com/rest/v3/SoftLayer_Virtual_Guest/"${vm_id}"?objectMask=primaryBackendIpAddress)
vm_private_ip=$(echo "${vm_private_ip_json}" | cut -d ":" -f2 | tr -d \"\{:\})

if [[ json -eq 1 ]]; then
    echo "${vm_private_ip_json}"
else
    echo "${vm_private_ip}"
fi