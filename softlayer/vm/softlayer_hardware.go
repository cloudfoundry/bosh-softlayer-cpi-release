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
	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
	datatypes "github.com/maximilien/softlayer-go/data_types"
)

type softLayerHardware struct {
	id int

	hardware datatypes.SoftLayer_Hardware

	softLayerClient sl.Client
	baremetalClient bmscl.BmpClient
	sshClient util.SshClient

	agentEnvService AgentEnvService

	logger boshlog.Logger
}

func NewSoftLayerHardware(id int, softLayerClient sl.Client, baremetalClient bmscl.BmpClient, sshClient util.SshClient, agentEnvService AgentEnvService, logger boshlog.Logger) VM {
	bslcommon.TIMEOUT = 60 * time.Minute
	bslcommon.POLLING_INTERVAL = 10 * time.Second

	hardware, err := bslcommon.GetObjectDetailsOnHardware(softLayerClient, id)
	if err != nil {
		return &softLayerHardware{}
	}

	return &softLayerHardware{
		id:       id,
		hardware: hardware,

		softLayerClient: softLayerClient,
		baremetalClient: baremetalClient,
		sshClient: sshClient,

		agentEnvService: agentEnvService,

		logger: logger,
	}
}

func (vm *softLayerHardware) ID() int { return vm.id }

func (vm *softLayerHardware) Delete(agentID string) error {
	updateStateResponse, err := vm.baremetalClient.UpdateState(strconv.Itoa(vm.ID()), "bm.state.deleted")
	if err != nil || updateStateResponse.Status != 200{
		return bosherr.WrapErrorf(err, "Faled to call bms to delete baremetal:"+string(body))
	}

	command := "rm -f /var/vcap/bosh/*.json ; sv stop agent"
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	return err
}

func (vm *softLayerHardware) Reboot() error {
	return nil
}

func (vm *softLayerHardware) ReloadOS(stemcell bslcstem.Stemcell) error {
	return nil
}

func (vm *softLayerHardware) SetMetadata(vmMetadata VMMetadata) error {
	return nil
}

func (vm *softLayerHardware) ConfigureNetworks(networks Networks) error {
	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", vm.ID())
	}

	oldAgentEnv.Networks = networks
	err = vm.agentEnvService.Update(oldAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring network setting on VirtualGuest with id: `%d`", vm.ID()))
	}

	return nil
}

func (vm *softLayerHardware) AttachDisk(disk bslcdisk.Disk) error {
	volume, err := vm.fetchIscsiVolume(disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d`", disk.ID()))
	}

	allowed, err := bslcommon.IscsiHasAllowedHardware(vm.softLayerClient, disk.ID(), vm.ID())

	totalTime := time.Duration(0)
	if err == nil && allowed == false {
		for totalTime < bslcommon.TIMEOUT {

			allowable, err := bslcommon.AttachHardwareIscsiVolume(vm.softLayerClient, vm.hardware, disk.ID())

			if err != nil {
				if !strings.Contains(err.Error(), "HTTP error code") {
					return bosherr.WrapError(err, fmt.Sprintf("Granting volume access to hardware %d", vm.ID()))
				}
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

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript()
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from virtual guest `%d`", vm.ID()))
	}

	deviceName, err := vm.waitForVolumeAttached(volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to attach volume `%d` to virtual guest `%d`", disk.ID(), vm.ID()))
	}
	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", vm.ID())
	}

	var newAgentEnv AgentEnv
	if hasMultiPath {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/mapper/"+deviceName)
	} else {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/"+deviceName)
	}

	err = vm.agentEnvService.Update(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on VirtualGuest with id: `%d`", vm.ID()))
	}

	return nil
}

func (vm *softLayerHardware) DetachDisk(disk bslcdisk.Disk) error {
	volume, err := vm.fetchIscsiVolume(disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("failed in disk `%d`", disk.ID()))
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript()
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from virtual guest `%d`", vm.ID()))
	}

	err = vm.detachVolumeBasedOnShellScript(vm.hardware, volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to detach volume with id %d from virtual guest with id: %d.", volume.Id, vm.ID())
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())
	if err == nil && allowed == true {
		//err = networkStorageService.DetachIscsiVolume(vm.hardware, disk.ID())
	}
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to revoke access of disk `%d` from virtual gusest `%d`", disk.ID(), vm.ID()))
	}

	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", vm.ID())
	}

	newAgentEnv := oldAgentEnv.DetachPersistentDisk(strconv.Itoa(disk.ID()))
	err = vm.agentEnvService.Update(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on VirtualGuest with id: `%d`", vm.ID()))
	}

	if len(newAgentEnv.Disks.Persistent) == 1 {
		for key, devicePath := range newAgentEnv.Disks.Persistent {
			leftDiskId, err := strconv.Atoi(key)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to transfer disk id %s from string to int", key))
			}
			vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "Left Disk Id %d", leftDiskId)
			vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "Left Disk device path %s", devicePath)
			volume, err := vm.fetchIscsiVolume(leftDiskId)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), vm.ID()))
			}

			_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(volume)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to virtual guest `%d`", key, vm.ID()))
			}

			command := fmt.Sprintf("sleep 5; mount %s-part1 /var/vcap/store", devicePath)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
			if err != nil {
				return bosherr.WrapError(err, "mount /var/vcap/store")
			}
		}
	}

	return nil
}

func (vm *softLayerHardware) GetDataCenterId() int {
	return vm.hardware.Datacenter.Id
}

func (vm *softLayerHardware) GetPrimaryIP() string {
	return vm.hardware.PrimaryIpAddress
}

func (vm *softLayerHardware) GetPrimaryBackendIP() string {
	return vm.hardware.PrimaryBackendIpAddress
}

func (vm *softLayerHardware) GetRootPassword() string {
	passwords := vm.hardware.OperatingSystem.Passwords
	for _, password := range passwords {
		if password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}
	return ""
}

func (vm *softLayerHardware) GetFullyQualifiedDomainName() string {
	return vm.hardware.FullyQualifiedDomainName
}

func (vm *softLayerHardware) SetVcapPassword(encryptedPwd string) (err error) {
	command := fmt.Sprintf("usermod -p '%s' vcap", encryptedPwd)
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to usermod vcap")
	}
	return
}

// Private methods
func (vm *softLayerHardware) waitForVolumeAttached(volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) (string, error) {

	oldDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(hasMultiPath)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from virtual guest `%d`", vm.ID()))
	}
	if len(oldDisks) > 2 {
		return "", bosherr.Error(fmt.Sprintf("Too manay persistent disks attached to virtual guest `%d`", vm.ID()))
	}

	credential, err := vm.getAllowedHostCredential()
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get iscsi host auth from virtual guest `%d`", vm.ID()))
	}

	_, err = vm.backupOpenIscsiConfBasedOnShellScript()
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to backup open iscsi conf files from virtual guest `%d`", vm.ID()))
	}

	_, err = vm.writeOpenIscsiInitiatornameBasedOnShellScript(credential)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi initiatorname from virtual guest `%d`", vm.ID()))
	}

	_, err = vm.writeOpenIscsiConfBasedOnShellScript(volume, credential)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi conf from virtual guest `%d`", vm.ID()))
	}

	_, err = vm.restartOpenIscsiBasedOnShellScript()
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to restart open iscsi from virtual guest `%d`", vm.ID()))
	}

	_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(volume)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to attach volume with id %d to virtual guest with id: %d.", volume.Id, vm.ID())
	}

	var deviceName string
	totalTime := time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {
		newDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(hasMultiPath)
		if err != nil {
			return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from virtual guest `%d`", vm.ID()))
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

	return "", bosherr.Errorf("Failed to attach disk '%d' to virtual guest '%d'", volume.Id, vm.ID())
}

func (vm *softLayerHardware) hasMulitPathToolBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("echo `command -v multipath`")
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	if err != nil {
		return false, err
	}

	if len(output) > 0 && strings.Contains(output, "multipath") {
		return true, nil
	}

	return false, nil
}

func (vm *softLayerHardware) getIscsiDeviceNamesBasedOnShellScript(hasMultiPath bool) ([]string, error) {
	devices := []string{}

	command1 := fmt.Sprintf("dmsetup ls")
	command2 := fmt.Sprintf("cat /proc/partitions")

	if hasMultiPath {
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command1)
		if err != nil {
			return devices, err
		}
		if strings.Contains(result, "No devices found") {
			return devices, nil
		}
		vm.logger.Info(SOFTLAYER_HARDWARE_LOG_TAG, fmt.Sprintf("Devices on hardware %d: %s", vm.ID(), result))
		lines := strings.Split(strings.Trim(result, "\n"), "\n")
		for i := 0; i < len(lines); i++ {
			if match, _ := regexp.MatchString("-part1", lines[i]); !match {
				devices = append(devices, strings.Fields(lines[i])[0])
			}
		}
	} else {
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command2)
		if err != nil {
			return devices, err
		}

		vm.logger.Info(SOFTLAYER_HARDWARE_LOG_TAG, fmt.Sprintf("Devices on hardware %d: %s", vm.ID(), result))
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

func (vm *softLayerHardware) fetchIscsiVolume(volumeId int) (datatypes.SoftLayer_Network_Storage, error) {
	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapError(err, "Cannot get network storage service.")
	}

	volume, err := networkStorageService.GetIscsiVolume(volumeId)
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Cannot get iSCSI volume with id: %d", volumeId)
	}

	return volume, nil
}

func (vm *softLayerHardware) getAllowedHostCredential() (AllowedHostCredential, error) {

	allowedHost, err := bslcommon.GetHardwareAllowedHost(vm.softLayerClient, vm.ID())

	if allowedHost.Id == 0 {
		return AllowedHostCredential{}, bosherr.Errorf("Cannot get allowed host with instance id: %d", vm.ID())
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

func (vm *softLayerHardware) backupOpenIscsiConfBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("cp /etc/iscsi/iscsid.conf{,.save}")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vm *softLayerHardware) restartOpenIscsiBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vm *softLayerHardware) discoveryOpenIscsiTargetsBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage) (bool, error) {
	command := fmt.Sprintf("sleep 5; iscsiadm -m discovery -t sendtargets -p %s", volume.ServiceResourceBackendIpAddress)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "discoverying open iscsi targets")
	}

	command = "sleep 5; echo `iscsiadm -m node -l`"
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "login iscsi targets")
	}

	return true, nil
}

func (vm *softLayerHardware) writeOpenIscsiInitiatornameBasedOnShellScript(credential AllowedHostCredential) (bool, error) {
	if len(credential.Iqn) > 0 {
		command := fmt.Sprintf("echo 'InitiatorName=%s' > /etc/iscsi/initiatorname.iscsi", credential.Iqn)
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
		if err != nil {
			return false, bosherr.WrapError(err, "Writing to /etc/iscsi/initiatorname.iscsi")
		}
	}

	return true, nil
}

func (vm *softLayerHardware) writeOpenIscsiConfBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage, credential AllowedHostCredential) (bool, error) {
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

	if err = vm.sshClient.UploadFile(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), file.Name(), "/etc/iscsi/iscsid.conf"); err != nil {
		return false, bosherr.WrapError(err, "Writing to /etc/iscsi/iscsid.conf")
	}

	return true, nil
}

func (vm *softLayerHardware) detachVolumeBasedOnShellScript(hardware datatypes.SoftLayer_Hardware, volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) error {
	// umount /var/vcap/store in case read-only mount
	isMounted, err := vm.isMountPoint(hardware, "/var/vcap/store")
	if err != nil {
		return bosherr.WrapError(err, "check mount point /var/vcap/store")
	}

	if isMounted {
		step00 := fmt.Sprintf("umount -l /var/vcap/store")
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), step00)
		if err != nil {
			return bosherr.WrapError(err, "umount -l /var/vcap/store")
		}
		vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "umount -l /var/vcap/store", nil)
	}

	// stop open-iscsi
	step1 := fmt.Sprintf("/etc/init.d/open-iscsi stop")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), step1)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "/etc/init.d/open-iscsi stop", nil)

	// clean up /etc/iscsi/send_targets/
	step2 := fmt.Sprintf("rm -rf /etc/iscsi/send_targets")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), step2)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets")
	}
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "rm -rf /etc/iscsi/send_targets", nil)

	// clean up /etc/iscsi/nodes/
	step3 := fmt.Sprintf("rm -rf /etc/iscsi/nodes")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), step3)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes")
	}

	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "rm -rf /etc/iscsi/nodes", nil)

	// start open-iscsi
	step4 := fmt.Sprintf("/etc/init.d/open-iscsi start")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), step4)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "/etc/init.d/open-iscsi start", nil)

	if hasMultiPath {
		// restart dm-multipath tool
		step5 := fmt.Sprintf("service multipath-tools restart")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), step5)
		if err != nil {
			return bosherr.WrapError(err, "Restarting Multipath deamon")
		}
		vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "service multipath-tools restart", nil)
	}

	return nil
}

func (vm *softLayerHardware) findOpenIscsiTargetBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest) ([]string, error) {
	command := "sleep 5 ; iscsiadm -m session -P3 | awk '/Target: /{print $2}'"
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
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

func (vm *softLayerHardware) findOpenIscsiPortalsBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage) ([]string, error) {
	command := "sleep 5 ; iscsiadm -m session -P3 | awk 'BEGIN{ lel=0} { if($0 ~ /Current Portal: /){ portal = $3 ; lel=NR } else { if( NR==(lel+46) && $0 ~ /Attached scsi disk /) {print portal}}}'"
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), command)
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

func (vm *softLayerHardware) postCheckActiveTransactionsForOSReload(softLayerClient sl.Client) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(vm.ID())
		if err != nil {
			if !strings.Contains(err.Error(), "HTTP error code") {
				return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
			}
		}

		if len(activeTransactions) > 0 {
			vm.logger.Info(SOFTLAYER_VM_OS_RELOAD_TAG, "OS Reload transaction started")
			break
		}

		totalTime += bslcommon.POLLING_INTERVAL
		time.Sleep(bslcommon.POLLING_INTERVAL)
	}

	if totalTime >= bslcommon.TIMEOUT {
		return errors.New(fmt.Sprintf("Waiting for OS Reload transaction to start TIME OUT!"))
	}

	err = bslcommon.WaitForVirtualGuest(vm.softLayerClient, vm.ID(), "RUNNING")
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id %d", vm.ID()))
		}
	}

	vm.logger.Info(SOFTLAYER_VM_OS_RELOAD_TAG, fmt.Sprintf("The virtual guest %d is powered on", vm.ID()))

	return nil
}

func (vm *softLayerHardware) isMountPoint(virtualGuest datatypes.SoftLayer_Hardware, path string) (bool, error) {
	mounts, err := vm.searchMounts(virtualGuest)
	if err != nil {
		return false, bosherr.WrapError(err, "Searching mounts")
	}

	for _, mount := range mounts {
		if mount.MountPoint == path {
			return true, nil
		}
	}

	return false, nil
}

func (vm *softLayerHardware) searchMounts(virtualGuest datatypes.SoftLayer_Hardware) ([]Mount, error) {
	var mounts []Mount
	stdout, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryIP(), "mount")
	if err != nil {
		return mounts, bosherr.WrapError(err, "Running mount")
	}

	// e.g. '/dev/sda on /boot type ext2 (rw)'
	for _, mountEntry := range strings.Split(stdout, "\n") {
		if mountEntry == "" {
			continue
		}

		mountFields := strings.Fields(mountEntry)

		mounts = append(mounts, Mount{
			PartitionPath: mountFields[0],
			MountPoint:    mountFields[2],
		})

	}

	return mounts, nil
}
