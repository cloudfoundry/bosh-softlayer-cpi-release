package vm

import (
	"bytes"
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

	sl "github.com/maximilien/softlayer-go/softlayer"

	bslcommon "github.com/maximilien/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"

	util "github.com/maximilien/bosh-softlayer-cpi/util"
	datatypes "github.com/maximilien/softlayer-go/data_types"
)

const (
	softLayerVMtag = "SoftLayerVM"
	ROOT_USER_NAME = "root"
	deleteVMLogTag = "DeleteVM"
)

type SoftLayerVM struct {
	id int

	softLayerClient sl.Client
	agentEnvService AgentEnvService

	sshClient util.SshClient

	logger boshlog.Logger
}

func NewSoftLayerVM(id int, softLayerClient sl.Client, sshClient util.SshClient, agentEnvService AgentEnvService, logger boshlog.Logger) SoftLayerVM {
	bslcommon.TIMEOUT = 60 * time.Minute
	bslcommon.POLLING_INTERVAL = 10 * time.Second

	return SoftLayerVM{
		id: id,

		softLayerClient: softLayerClient,
		agentEnvService: agentEnvService,

		sshClient: sshClient,

		logger: logger,
	}
}

func (vm SoftLayerVM) ID() int { return vm.id }

func (vm SoftLayerVM) Delete() error {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating SoftLayer VirtualGuestService from client")
	}

	vmCID := vm.ID()
	err = bslcommon.WaitForVirtualGuestToHaveNoRunningTransactions(vm.softLayerClient, vmCID)
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

	err = vm.postCheckActiveTransactionsForDeleteVM(vm.softLayerClient, vmCID)
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
	virtualGuest, err := bslcommon.GetObjectDetailsOnVirtualGuest(vm.softLayerClient, vm.ID())
	if err != nil {
		return bosherr.WrapErrorf(err, "Cannot get details from virtual guest with id: %d.", virtualGuest.Id)
	}

	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", virtualGuest.Id)
	}

	oldAgentEnv.Networks = networks
	err = vm.agentEnvService.Update(oldAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring network setting on VirtualGuest with id: `%d`", virtualGuest.Id))
	}

	return nil
}

func (vm SoftLayerVM) AttachDisk(disk bslcdisk.Disk) error {
	virtualGuest, volume, err := vm.fetchVMandIscsiVolume(vm.ID(), disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), virtualGuest.Id))
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())

	totalTime := time.Duration(0)
	if err == nil && allowed == false {
		for totalTime < bslcommon.TIMEOUT {
			allowable, err := networkStorageService.AttachIscsiVolume(virtualGuest, disk.ID())
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Granting volume access to vitrual guest %d", virtualGuest.Id))
			} else {

				if allowable {
					break
				}
			}

			totalTime += bslcommon.POLLING_INTERVAL
			time.Sleep(bslcommon.POLLING_INTERVAL)
		}
	}
	if totalTime >= bslcommon.TIMEOUT {
		return bosherr.Error("Waiting for grantting access to virutal guest TIME OUT!")
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript(virtualGuest)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from virtual guest `%d`", virtualGuest.Id))
	}

	deviceName, err := vm.waitForVolumeAttached(virtualGuest, volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to attach volume `%d` to virtual guest `%d`", disk.ID(), virtualGuest.Id))
	}

	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", virtualGuest.Id)
	}

	var newAgentEnv AgentEnv
	if hasMultiPath {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/mapper/"+deviceName)
	} else {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/"+deviceName)
	}

	err = vm.agentEnvService.Update(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on VirtualGuest with id: `%d`", virtualGuest.Id))
	}

	return nil
}

func (vm SoftLayerVM) DetachDisk(disk bslcdisk.Disk) error {
	virtualGuest, volume, err := vm.fetchVMandIscsiVolume(vm.ID(), disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("failed in disk `%d` from virtual gusest `%d`", disk.ID(), virtualGuest.Id))
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript(virtualGuest)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from virtual guest `%d`", virtualGuest.Id))
	}

	err = vm.detachVolumeBasedOnShellScript(virtualGuest, volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to detach volume with id %d from virtual guest with id: %d.", volume.Id, virtualGuest.Id)
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())
	if err == nil && allowed == true {
		err = networkStorageService.DetachIscsiVolume(virtualGuest, disk.ID())
	}
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to revoke access of disk `%d` from virtual gusest `%d`", disk.ID(), virtualGuest.Id))
	}

	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", virtualGuest.Id)
	}

	newAgentEnv := oldAgentEnv.DetachPersistentDisk(strconv.Itoa(disk.ID()))
	err = vm.agentEnvService.Update(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on VirtualGuest with id: `%d`", virtualGuest.Id))
	}

	if len(newAgentEnv.Disks.Persistent) == 1 {
		for key, _ := range newAgentEnv.Disks.Persistent {
			existingDiskId, err := strconv.Atoi(key)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to transfer disk id %s from string to int", key))
			}

			virtualGuest, volume, err := vm.fetchVMandIscsiVolume(vm.ID(), existingDiskId)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), virtualGuest.Id))
			}

			networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
			if err != nil {
				return bosherr.WrapError(err, "Cannot get network storage service.")
			}

			allowed, err := networkStorageService.HasAllowedVirtualGuest(existingDiskId, vm.ID())

			totalTime := time.Duration(0)
			if err == nil && allowed == false {
				for totalTime < bslcommon.TIMEOUT {
					allowable, err := networkStorageService.AttachIscsiVolume(virtualGuest, existingDiskId)
					if err != nil {
						return bosherr.WrapError(err, fmt.Sprintf("Granting volume access to vitrual guest %d", virtualGuest.Id))
					} else {

						if allowable {
							break
						}
					}
					totalTime += bslcommon.POLLING_INTERVAL
					time.Sleep(bslcommon.POLLING_INTERVAL)
				}
			}
			if totalTime >= bslcommon.TIMEOUT {
				return bosherr.Error("Waiting for grantting access to virutal guest TIME OUT!")
			}

			deviceName, err := vm.waitForVolumeAttached(virtualGuest, volume, hasMultiPath)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to virtual guest `%d`", key, virtualGuest.Id))
			}

			if len(deviceName) > 0 {
				return nil
			} else {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to virtual guest `%d`", key, virtualGuest.Id))
			}
		}
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
				return []string{}, bosherr.Errorf("Cannot convert tags metadata value `%v` to string", value)
			}

			tags = vm.parseTags(stringValue)
		}
	}

	return tags, nil
}

func (vm SoftLayerVM) parseTags(value string) []string {
	return strings.Split(value, ",")
}

func (vm SoftLayerVM) waitForVolumeAttached(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) (string, error) {
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
	for totalTime < bslcommon.TIMEOUT {
		if _, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(virtualGuest, volume); err != nil {
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

		totalTime += bslcommon.POLLING_INTERVAL
		time.Sleep(bslcommon.POLLING_INTERVAL)
	}

	return "", bosherr.Errorf("Failed to attach disk '%d' to virtual guest '%d'", volume.Id, virtualGuest.Id)
}

func (vm SoftLayerVM) hasMulitPathToolBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := fmt.Sprintf("echo `command -v multipath`")
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, err
	}

	if len(output) > 0 && strings.Contains(output, "multipath") {
		return true, nil
	}
	return false, nil
}

func (vm SoftLayerVM) getIscsiDeviceNamesBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, hasMultiPath bool) ([]string, error) {
	devices := []string{}

	command1 := fmt.Sprintf("dmsetup ls")
	command2 := fmt.Sprintf("cat /proc/partitions")

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
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Cannot get softlayer virtual guest service.")
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Cannot get network storage service.")
	}

	virtualGuest, err := virtualGuestService.GetObject(vmId)
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Cannot get virtual guest with id: %d", vmId)
	}

	volume, err := networkStorageService.GetIscsiVolume(volumeId)
	if err != nil {
		return datatypes.SoftLayer_Virtual_Guest{}, datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Cannot get iSCSI volume with id: %d", volumeId)
	}

	return virtualGuest, volume, nil
}

func (vm SoftLayerVM) getAllowedHostCredential(virtualGuest datatypes.SoftLayer_Virtual_Guest) (AllowedHostCredential, error) {
	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapError(err, "Cannot get softlayer virtual guest service.")
	}

	allowedHost, err := virtualGuestService.GetAllowedHost(virtualGuest.Id)
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapErrorf(err, "Cannot get allowed host with instance id: %d", virtualGuest.Id)
	}
	if allowedHost.Id == 0 {
		return AllowedHostCredential{}, bosherr.Errorf("Cannot get allowed host with instance id: %d", virtualGuest.Id)
	}

	allowedHostService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Allowed_Host_Service()
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapError(err, "Cannot get network storage allowed host service.")
	}

	credential, err := allowedHostService.GetCredential(allowedHost.Id)
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapErrorf(err, "Cannot get credential with allowed host id: %d", allowedHost.Id)
	}

	return AllowedHostCredential{
		Iqn:      allowedHost.Name,
		Username: credential.Username,
		Password: credential.Password,
	}, nil
}

func (vm SoftLayerVM) backupOpenIscsiConfBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := fmt.Sprintf("cp /etc/iscsi/iscsid.conf{,.save}")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vm SoftLayerVM) restartOpenIscsiBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) (bool, error) {
	command := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vm SoftLayerVM) discoveryOpenIscsiTargetsBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage) (bool, error) {
	command := fmt.Sprintf("sleep 5; iscsiadm -m discoverydb -t sendtargets -p %s -o new -o delete --discover", volume.ServiceResourceBackendIpAddress)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "discvoerying open iscsi targets")
	}

	command = "sleep 5; echo `iscsiadm -m node -l`"
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "login iscsi targets")
	}

	return true, nil
}

func (vm SoftLayerVM) writeOpenIscsiInitiatornameBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, credential AllowedHostCredential) (bool, error) {
	if len(credential.Iqn) > 0 {
		command := fmt.Sprintf("echo 'InitiatorName=%s' > /etc/iscsi/initiatorname.iscsi", credential.Iqn)
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
node.conn[0].timeo.noop_out_interval = 5
node.conn[0].timeo.noop_out_timeout = 10
`

func (vm SoftLayerVM) detachVolumeBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) error {
	targets, err := vm.findOpenIscsiTargetBasedOnShellScript(virtualGuest)
	if err != nil {
		return err
	}

	if len(targets) == 1 {
		portals, err := vm.findOpenIscsiPortalsBasedOnShellScript(virtualGuest, volume)
		if err != nil {
			return err
		}

		for _, portal := range portals {
			step1 := fmt.Sprintf("iscsiadm -m node -T %s --portal %s:3260 -u", targets[0], portal)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step1)
			if err != nil {
				return bosherr.WrapErrorf(err, "Logout portal: %s", portal)
			}

			step2 := fmt.Sprintf("iscsiadm -m node -o delete -T %s:3260 --portal %s", targets[0], portal)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step2)
			if err != nil {
				return bosherr.WrapErrorf(err, "Removing iSCSI portal: %s", portal)
			}

			step3 := fmt.Sprintf("iscsiadm -m discoverydb -t sendtargets -p %s:3260 -o delete", portal)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step3)
			if err != nil {
				return bosherr.WrapErrorf(err, "Deleting discovery record from portal: %s", portal)
			}
		}
	} else {
		step1 := fmt.Sprintf("iscsiadm -m node -u")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step1)
		if err != nil {
			return bosherr.WrapErrorf(err, "Logout all portals")
		}
	}

	// clean up /etc/iscsi/send_targets/
	step4 := fmt.Sprintf("rm -r /etc/iscsi/send_targets/")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step4)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets/")
	}
	// clean up /etc/iscsi/nodes/
	step5 := fmt.Sprintf("rm -r /etc/iscsi/nodes/")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step5)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes/")
	}
	// restart open-iscsi
	step6 := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step6)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}

	if hasMultiPath {
		// restart dm-multipath tool
		step7 := fmt.Sprintf("service multipath-tools restart")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step7)
		if err != nil {
			return bosherr.WrapError(err, "Restarting Multipath deamon")
		}
	}

	return nil
}

func (vm SoftLayerVM) findOpenIscsiTargetBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) ([]string, error) {
	command := "sleep 5 ; iscsiadm -m session -P3 | awk '/Target: /{print $2}'"
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return []string{}, err
	}

	targets := []string{}
	lines := strings.Split(strings.Trim(output, "\n"), "\n")
	for _, line := range lines {
		targets = append(targets, strings.Split(line, ",")[0])
	}

	if len(targets) > 0 {
		return targets, nil
	}

	return []string{}, errors.New(fmt.Sprintf("Cannot find matched iSCSI device"))
}

func (vm SoftLayerVM) findOpenIscsiPortalsBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage) ([]string, error) {
	command := "sleep 5 ; iscsiadm -m session -P3 | awk 'BEGIN{ lel=0} { if($0 ~ /Current Portal: /){ portal = $3 ; lel=NR } else { if( NR==(lel+46) && $0 ~ /Attached scsi disk /) {print portal}}}'"
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

func (vm SoftLayerVM) postCheckActiveTransactionsForDeleteVM(softLayerClient sl.Client, virtualGuestId int) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
		}

		if len(activeTransactions) > 0 {
			vm.logger.Info(deleteVMLogTag, "Delete VM transaction started", nil)
			break
		}

		totalTime += bslcommon.POLLING_INTERVAL
		time.Sleep(bslcommon.POLLING_INTERVAL)
	}

	if totalTime >= bslcommon.TIMEOUT {
		return errors.New(fmt.Sprintf("Waiting for DeleteVM transaction to start TIME OUT!"))
	}

	totalTime = time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {
		vm1, err := virtualGuestService.GetObject(virtualGuestId)
		if err != nil || vm1.Id == 0 {
			vm.logger.Info(deleteVMLogTag, "VM doesn't exist. Delete done", nil)
			break
		}

		activeTransaction, err := virtualGuestService.GetActiveTransaction(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
		}
		
		averageDuration := activeTransaction.TransactionStatus.AverageDuration
		if strings.HasPrefix(averageDuration, ".") {
			averageDuration = "0" + averageDuration
		}

		averageTransactionDuration, err := strconv.ParseFloat(averageDuration, 32)
		if err != nil {
			return bosherr.WrapError(err, "Parsing float for average transaction duration")
		}

		if averageTransactionDuration > 30 {
			vm.logger.Info(deleteVMLogTag, "Deleting VM instance had been launched and it is a long transaction. Please check Softlayer Portal", nil)
			break
		}

		vm.logger.Info(deleteVMLogTag, "This is a short transaction, waiting for all active transactions to complete", nil)
		totalTime += bslcommon.POLLING_INTERVAL
		time.Sleep(bslcommon.POLLING_INTERVAL)
	}

	if totalTime >= bslcommon.TIMEOUT {
		return errors.New(fmt.Sprintf("After deleting a vm, waiting for active transactions to complete TIME OUT!"))
	}

	return nil
}

func (vm SoftLayerVM) execCommand(virtualGuest datatypes.SoftLayer_Virtual_Guest, command string) (string, error) {
	result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	return result, err
}

func (vm SoftLayerVM) uploadFile(virtualGuest datatypes.SoftLayer_Virtual_Guest, srcFile string, destFile string) error {
	err := vm.sshClient.UploadFile(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, srcFile, destFile)
	return err
}

func (vm SoftLayerVM) downloadFile(virtualGuest datatypes.SoftLayer_Virtual_Guest, srcFile string, destFile string) error {
	err := vm.sshClient.DownloadFile(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, srcFile, destFile)
	return err
}
