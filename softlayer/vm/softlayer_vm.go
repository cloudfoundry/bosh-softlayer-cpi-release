package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	"github.com/pivotal-golang/clock"

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
	datatypes "github.com/maximilien/softlayer-go/data_types"
)

const (
	softLayerVMtag                 = "SoftLayerVM"
	ROOT_USER_NAME                 = "root"
	deleteVMLogTag                 = "DeleteVM"
	TIMEOUT_TRANSACTIONS_DELETE_VM = 60 * time.Minute
)

type SoftLayerVM struct {
	id int

	softLayerClient sl.Client
	agentEnvService AgentEnvService

	sshClient util.SshClient

	logger boshlog.Logger

	timeoutForActiveTransactions time.Duration
}

func NewSoftLayerVM(id int, softLayerClient sl.Client, sshClient util.SshClient, agentEnvService AgentEnvService, logger boshlog.Logger, timeoutForActiveTransactions time.Duration) SoftLayerVM {
	bslcommon.TIMEOUT = 10 * time.Minute
	bslcommon.POLLING_INTERVAL = 10 * time.Second

	return SoftLayerVM{
		id: id,

		softLayerClient: softLayerClient,
		agentEnvService: agentEnvService,

		sshClient: sshClient,

		logger: logger,

		timeoutForActiveTransactions: timeoutForActiveTransactions,
	}
}

func (vm SoftLayerVM) ID() int { return vm.id }

func (vm SoftLayerVM) Delete() error {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	vmCID := vm.ID()
	err = bslcommon.WaitForVirtualGuestToHaveNoRunningTransactions(vm.softLayerClient, vmCID, vm.timeoutForActiveTransactions, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest `%d` to have no pending transactions before deleting vm", vmCID))
	}

	deleted, err := virtualGuestService.DeleteObject(vm.ID())
	if err != nil {
		return bosherr.WrapError(err, "Deleting SoftLayer VirtualGuest from client")
	}

	if !deleted {
		return bosherr.WrapError(nil, "Did not delete SoftLayer VirtualGuest from client")
	}

	err = vm.postCheckActiveTransactionsForDeleteVM(vm.softLayerClient, vmCID, vm.timeoutForActiveTransactions, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return err
	}

	return nil
}

func (vm SoftLayerVM) Reboot() error {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	rebooted, err := virtualGuestService.RebootSoft(vm.ID())
	if err != nil {
		return bosherr.WrapError(err, "Rebooting (soft) SoftLayer VirtualGuest from client")
	}

	if !rebooted {
		return bosherr.WrapError(nil, "Did not reboot (soft) SoftLayer VirtualGuest from client")
	}

	return nil
}

func (vm SoftLayerVM) SetMetadata(vmMetadata VMMetadata) error {
	tags, err := vm.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	//Check below needed since Golang strings.Split return [""] on strings.Split("", ",")
	if len(tags) == 1 && tags[0] == "" {
		return nil
	}

	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	success, err := virtualGuestService.SetTags(vm.ID(), tags)
	if !success {
		return bosherr.WrapErrorf(err, "Settings tags on SoftLayer VirtualGuest `%d`", vm.ID())
	}

	if err != nil {
		return bosherr.WrapErrorf(err, "Settings tags on SoftLayer VirtualGuest `%d`", vm.ID())
	}

	return nil
}

func (vm SoftLayerVM) ConfigureNetworks(networks Networks) error {
	return NotSupportedError{}
}

func (vm SoftLayerVM) AttachDisk(disk bslcdisk.Disk) error {
	virtualGuest, volume, err := vm.fetchVMandIscsiVolume(vm.ID(), disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), virtualGuest.Id))
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Can not get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())
	if err == nil && allowed == false {
		attachIscsiVolumeRetryable := boshretry.NewRetryable(
			func() (bool, error) {
				resp, err := networkStorageService.AttachIscsiVolume(virtualGuest, disk.ID())
				if err != nil || strings.Contains(resp, "A Volume Provisioning is currently in progress") {
					return true, bosherr.WrapError(err, fmt.Sprintf("Granting volume access to vitrual guest %d", virtualGuest.Id))
				}
				return false, nil
			})
		timeService := clock.NewClock()
		timeoutRetryStrategy := boshretry.NewTimeoutRetryStrategy(bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL, attachIscsiVolumeRetryable, timeService, vm.logger)
		err := timeoutRetryStrategy.Try()
		if err != nil {
			return bosherr.WrapError(err, fmt.Sprintf("Failed to grant access of disk `%d` from virtual guest `%d`", disk.ID(), virtualGuest.Id))
		}
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript(virtualGuest)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from virtual guest `%d`", virtualGuest.Id))
	}

	deviceName, err := vm.waitForVolumeAttached(virtualGuest, volume, hasMultiPath, bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to attach volume `%d` to virtual guest `%d`", disk.ID(), virtualGuest.Id))
	}

	metadata, err := bslcommon.GetUserMetadataOnVirtualGuest(vm.softLayerClient, virtualGuest.Id)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to get metadata from virtual guest with id: %d.", virtualGuest.Id)
	}

	oldAgentEnv, err := NewAgentEnvFromJSON(metadata)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal metadata from virutal guest with id: %d.", virtualGuest.Id)
	}

	var newAgentEnv AgentEnv
	if hasMultiPath {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/mapper/"+deviceName)
	} else {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/"+deviceName)
	}

	metadata, err = json.Marshal(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent environment metadata")
	}

	err = bslcommon.ConfigureMetadataOnVirtualGuest(vm.softLayerClient, virtualGuest.Id, string(metadata), bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuest.Id))
	}

	return nil
}

func (vm SoftLayerVM) DetachDisk(disk bslcdisk.Disk) error {
	virtualGuest, volume, err := vm.fetchVMandIscsiVolume(vm.ID(), disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("failed in disk `%d` from virtual gusest `%d`", disk.ID(), virtualGuest.Id))
	}

	err = vm.detachVolumeBasedOnShellScript(virtualGuest, volume)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to detach volume with id %d from virtual guest with id: %d.", volume.Id, virtualGuest.Id)
	}

	metadata, err := bslcommon.GetUserMetadataOnVirtualGuest(vm.softLayerClient, virtualGuest.Id)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to get metadata from virtual guest with id: %d.", virtualGuest.Id)
	}
	oldAgentEnv, err := NewAgentEnvFromJSON(metadata)
	newAgentEnv := oldAgentEnv.DetachPersistentDisk(strconv.Itoa(disk.ID()))

	metadata, err = json.Marshal(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent environment metadata")
	}

	err = bslcommon.ConfigureMetadataOnVirtualGuest(vm.softLayerClient, virtualGuest.Id, string(metadata), bslcommon.TIMEOUT, bslcommon.POLLING_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring metadata on VirtualGuest `%d`", virtualGuest.Id))
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Can not get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())
	if err == nil && allowed == true {
		err = networkStorageService.DetachIscsiVolume(virtualGuest, disk.ID())
	}
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to revoke access of disk `%d` from virtual gusest `%d`", disk.ID(), virtualGuest.Id))
	}

	return nil
}

// Private methods

func (vm SoftLayerVM) extractTagsFromVMMetadata(vmMetadata VMMetadata) ([]string, error) {
	tags := []string{}
	for key, value := range vmMetadata {
		if key == "tags" {
			stringValue, ok := value.(string)
			if !ok {
				return []string{}, bosherr.Errorf("Could not convert tags metadata value `%v` to string", value)
			}

			tags = vm.parseTags(stringValue)
		}
	}

	return tags, nil
}

func (vm SoftLayerVM) parseTags(value string) []string {
	return strings.Split(value, ",")
}

func (vm SoftLayerVM) waitForVolumeAttached(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool, timeout, pollingInterval time.Duration) (string, error) {
	var deviceName string

	oldDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(virtualGuest, hasMultiPath)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from virtual guest `%d`", virtualGuest.Id))
	}
	if len(oldDisks) > 2 {
		return "", bosherr.Error(fmt.Sprintf("Too manay persistent disks attached to virtual guest `%d`", virtualGuest.Id))
	}

	credential, err := vm.getAllowedHostCredential(virtualGuest)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get iscsi host auth from virtual guest `%d`", virtualGuest.Id))
	}

	if _, err = vm.backupOpenIscsiConfBasedOnShellScript(virtualGuest); err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to backup open iscsi conf files from virtual guest `%d`", virtualGuest.Id))
	}

	if _, err = vm.writeOpenIscsiInitiatornameBasedOnShellScript(virtualGuest, credential); err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi initiatorname from virtual guest `%d`", virtualGuest.Id))
	}

	if _, err = vm.writeOpenIscsiConfBasedOnShellScript(virtualGuest, volume, credential); err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi conf from virtual guest `%d`", virtualGuest.Id))
	}

	if _, err = vm.restartOpenIscsiBasedOnShellScript(virtualGuest); err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to restart open iscsi from virtual guest `%d`", virtualGuest.Id))
	}

	totalTime := time.Duration(0)
	for totalTime < timeout {
		if _, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(virtualGuest); err != nil {
			return "", bosherr.WrapErrorf(err, "Failed to attach volume with id %d to virtual guest with id: %d.", volume.Id, virtualGuest.Id)
		}
		newDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(virtualGuest, hasMultiPath)
		if err != nil {
			return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from virtual guest `%d`", virtualGuest.Id))
		}

		if len(oldDisks) == 0 {
			if len(newDisks) > 0 {
				deviceName = newDisks[0]
				return deviceName, nil
			}
		}

		var included bool
		for _, newDisk := range newDisks {
			for _, oldDisk := range oldDisks {
				if strings.EqualFold(newDisk, oldDisk) {
					included = true
				}
			}
			if !included {
				deviceName = newDisk
			}
			included = false
		}

		if len(deviceName) > 0 {
			return deviceName, nil
		}

		totalTime += pollingInterval
		time.Sleep(pollingInterval)
	}

	return "", bosherr.Errorf("Failed to attach disk '%d' to virtual guest '%d'", volume.Id, virtualGuest.Id)
}

func (vm SoftLayerVM) hasMulitPathToolBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := `command -v multipath`
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, err
	}

	if strings.Contains(output, "multipath") {
		return true, nil
	}
	return false, nil
}

func (vm SoftLayerVM) getIscsiDeviceNamesBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, hasMultiPath bool) ([]string, error) {
	devices := []string{}

	command1 := `dmsetup ls`
	command2 := `cat /proc/partitions`

	if hasMultiPath {
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command1)
		if err != nil {
			return devices, err
		}
		if strings.Contains(result, "No devices found") {
			return devices, nil
		}

		lines := strings.Split(strings.Trim(result, "\n"), "\n")
		for i := 0; i < len(lines); i++ {
			if match, _ := regexp.MatchString("-part1", lines[i]); !match {
				devices = append(devices, strings.Fields(lines[i])[0])
			}
		}
	} else {
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command2)
		if err != nil {
			return devices, err
		}

		lines := strings.Split(strings.Trim(result, "\n"), "\n")
		for i := 0; i < len(lines); i++ {
			if match, _ := regexp.MatchString("sd[a-z]$", lines[i]); match {
				vals := strings.Fields(lines[i])
				devices = append(devices, vals[len(vals)-1])
			}
		}
	}
	return devices, nil
}

func (vm SoftLayerVM) fetchVMandIscsiVolume(vmId int, volumeId int) (datatypes.SoftLayer_Virtual_Guest, datatypes.SoftLayer_Network_Storage, error) {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Can not get softlayer virtual guest service.")
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Can not get network storage service.")
	}

	virtualGuest, err := virtualGuestService.GetObject(vmId)
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Can not get virtual guest with id: %d", vmId)
	}

	volume, err := networkStorageService.GetIscsiVolume(volumeId)
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Failed to get iSCSI volume with id: %d", volumeId)
	}

	return virtualGuest, volume, nil
}

func (vm SoftLayerVM) getAllowedHostCredential(virtualGuest datatypes.SoftLayer_Virtual_Guest) (AllowedHostCredential, error) {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapError(err, "Can not get softlayer virtual guest service.")
	}

	allowedHost, err := virtualGuestService.GetAllowedHost(virtualGuest.Id)
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapErrorf(err, "Can not get allowed host with instance id: %d", virtualGuest.Id)
	}

	allowedHostService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Allowed_Host_Service()
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapError(err, "Can not get network storage allowed host service.")
	}

	credential, err := allowedHostService.GetCredential(allowedHost.Id)
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapErrorf(err, "Can not get credential with allowed host id: %d", allowedHost.Id)
	}

	return AllowedHostCredential{
		Iqn:      allowedHost.Name,
		Username: credential.Username,
		Password: credential.Password,
	}, nil
}

func (vm SoftLayerVM) backupOpenIscsiConfBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := fmt.Sprintf(`
						cp /etc/iscsi/iscsid.conf{,.save}`,
	)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vm SoftLayerVM) restartOpenIscsiBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := fmt.Sprintf(`
						export PATH=/etc/init.d:$PATH
						open-iscsi restart`,
	)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vm SoftLayerVM) discoveryOpenIscsiTargetsBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := fmt.Sprintf(`
						iscsiadm -m discovery -t sendtargets -p %s
						iscsiadm -m node -l`,
		virtualGuest.PrimaryBackendIpAddress,
	)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "discvoerying open iscsi targets")
	}

	return true, nil
}

func (vm SoftLayerVM) writeOpenIscsiInitiatornameBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, credential AllowedHostCredential) (bool, error) {
	if len(credential.Iqn) > 0 {
		command := fmt.Sprintf(`
						echo "InitiatorName=%s" > /etc/iscsi/initiatorname.iscsi`,
			credential.Iqn,
		)
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
		if err != nil {
			return false, bosherr.WrapError(err, "Writing to /etc/iscsi/initiatorname.iscsi")
		}
	}

	return true, nil
}

func (vm SoftLayerVM) writeOpenIscsiConfBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage, credential AllowedHostCredential) (bool, error) {
	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("open_iscsid_conf").Parse(etcIscsidConfTemplate))
	if len(credential.Password) == 0 {
		err := t.Execute(buffer, volume)
		if err != nil {
			return false, bosherr.WrapError(err, "Generating config from template")
		}
	} else {
		err := t.Execute(buffer, credential)
		if err != nil {
			return false, bosherr.WrapError(err, "Generating config from template")
		}
	}

	file, err := ioutil.TempFile(os.TempDir(), "iscsid_conf_")
	if err != nil {
		return false, bosherr.WrapError(err, "Generating config from template")
	}

	defer os.Remove(file.Name())

	_, err = file.WriteString(buffer.String())
	if err != nil {
		return false, bosherr.WrapError(err, "Generating config from template")
	}

	if err = vm.uploadFile(virtualGuest, file.Name(), "/etc/iscsi/iscsid.conf"); err != nil {
		return false, bosherr.WrapError(err, "Writing to /etc/iscsi/iscsid.conf")
	}

	return true, nil
}

const etcIscsidConfTemplate = `# Generated by bosh-agent
node.startup = automatic
node.session.auth.authmethod = CHAP
node.session.auth.username = {{.Username}}
node.session.auth.password = {{.Password}}
discovery.sendtargets.auth.authmethod = CHAP
discovery.sendtargets.auth.username = {{.Username}}
discovery.sendtargets.auth.password = {{.Password}}
node.session.timeo.replacement_timeout = 120
node.conn[0].timeo.login_timeout = 15
node.conn[0].timeo.logout_timeout = 15
node.conn[0].timeo.noop_out_interval = 10
node.conn[0].timeo.noop_out_timeout = 15
node.session.iscsi.InitialR2T = No
node.session.iscsi.ImmediateData = Yes
node.session.iscsi.FirstBurstLength = 262144
node.session.iscsi.MaxBurstLength = 16776192
node.conn[0].iscsi.MaxRecvDataSegmentLength = 65536
`

func (vm SoftLayerVM) detachVolumeBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage) error {
	targetName, err := vm.findOpenIscsiTargetBasedOnShellScript(virtualGuest, volume)
	if err != nil {
		return err
	}

	portals, err := vm.findOpenIscsiPortalsBasedOnShellScript(virtualGuest, volume)
	if err != nil {
		return err
	}

	for _, portal := range portals {
		command := fmt.Sprintf(`
		iscsiadm -m node -T %s --portal %s -u
		iscsiadm -m node -o delete -T %s --portal %s`,
			targetName,
			portal,
			targetName,
			portal,
		)

		if _, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command); err != nil {
			return err
		}
	}

	return nil
}

func (vm SoftLayerVM) findOpenIscsiTargetBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage) (targetName string, err error) {
	command := `
		sleep 1
			iscsiadm -m session -P3 | awk '/Target: /{print $2}'`

	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.Trim(output, "\n"), "\n")
	for _, line := range lines {
		return line, nil
	}

	return "", errors.New(fmt.Sprintf("Can not find matched iSCSI device for user name: %s", volume.Username))
}

func (vm SoftLayerVM) findOpenIscsiPortalsBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage) ([]string, error) {
	command := `
		sleep 1
			iscsiadm -m session -P3 | awk 'BEGIN{ lel=0} { if($0 ~ /Current Portal: /){ portal = $3 ; lel=NR } else { if( NR==(lel+46) && $0 ~ /Attached scsi disk /) {print portal}}}'`
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return []string{}, err
	}

	portals := []string{}
	lines := strings.Split(strings.Trim(output, "\n"), "\n")
	for _, line := range lines {
		portals = append(portals, strings.Split(line, ",")[0])
	}
	return portals, nil
}

func (vm SoftLayerVM) getRootPassword(virtualGuest datatypes.SoftLayer_Virtual_Guest) string {
	passwords := virtualGuest.OperatingSystem.Passwords

	for _, password := range passwords {
		if password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}

	return ""
}

func (vm SoftLayerVM) postCheckActiveTransactionsForDeleteVM(softLayerClient sl.Client, virtualGuestId int, timeout, pollingInterval time.Duration) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < timeout {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
		}

		if len(activeTransactions) > 0 {
			vm.logger.Info(deleteVMLogTag, "Delete VM transaction started", nil)
			break
		}

		totalTime += pollingInterval
		time.Sleep(pollingInterval)
	}

	if totalTime >= timeout {
		return errors.New(fmt.Sprintf("Waiting for DeleteVM transaction to start TIME OUT!"))
	}

	totalTime = time.Duration(0)
	for totalTime < timeout {
		vm1, err := virtualGuestService.GetObject(virtualGuestId)
		if err != nil || vm1.Id == 0 {
			vm.logger.Info(deleteVMLogTag, "VM doesn't exist. Delete done", nil)
			break
		}

		activeTransaction, err := virtualGuestService.GetActiveTransaction(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
		}

		averageTransactionDuration, err := strconv.ParseFloat(activeTransaction.TransactionStatus.AverageDuration, 32)
		if err != nil {
			return bosherr.WrapError(err, "Parsing float for average transaction duration")
		}

		if averageTransactionDuration > 30 {
			vm.logger.Info(deleteVMLogTag, "Deleting VM instance had been launched and it is a long transaction. Please check Softlayer Portal", nil)
			break
		}

		vm.logger.Info(deleteVMLogTag, "This is a short transaction, waiting for all active transactions to complete", nil)
		totalTime += pollingInterval
		time.Sleep(pollingInterval)
	}

	if totalTime >= timeout {
		return errors.New(fmt.Sprintf("After deleting a vm, waiting for active transactions to complete TIME OUT!"))
	}

	return nil
}

func (vm SoftLayerVM) execCommand(virtualGuest datatypes.SoftLayer_Virtual_Guest, command string) (string, error) {
	result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryIpAddress, command)
	return result, err
}

func (vm SoftLayerVM) uploadFile(virtualGuest datatypes.SoftLayer_Virtual_Guest, srcFile string, destFile string) error {
	err := vm.sshClient.UploadFile(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryIpAddress, srcFile, destFile)
	return err
}

func (vm SoftLayerVM) downloadFile(virtualGuest datatypes.SoftLayer_Virtual_Guest, srcFile string, destFile string) error {
	err := vm.sshClient.DownloadFile(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryIpAddress, srcFile, destFile)
	return err
}
