package hardware

import (
	"bytes"
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

	bmscl "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	sl "github.com/maximilien/softlayer-go/softlayer"

	"github.com/cloudfoundry/bosh-softlayer-cpi/api"
	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	"github.com/cloudfoundry/bosh-softlayer-cpi/util"

	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	datatypes "github.com/maximilien/softlayer-go/data_types"
)

type softLayerHardware struct {
	id int

	hardware datatypes.SoftLayer_Hardware

	softLayerClient sl.Client
	baremetalClient bmscl.BmpClient
	sshClient       util.SshClient

	agentEnvService AgentEnvService

	logger boshlog.Logger
}

func NewSoftLayerHardware(hardware datatypes.SoftLayer_Hardware, softLayerClient sl.Client, baremetalClient bmscl.BmpClient, sshClient util.SshClient, logger boshlog.Logger) VM {
	slh.TIMEOUT = 60 * time.Minute
	slh.POLLING_INTERVAL = 10 * time.Second

	return &softLayerHardware{
		id: hardware.Id,

		hardware: hardware,

		softLayerClient: softLayerClient,
		baremetalClient: baremetalClient,
		sshClient:       sshClient,

		logger: logger,
	}
}

func (vm *softLayerHardware) ID() int { return vm.id }

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

func (vm *softLayerHardware) SetAgentEnvService(agentEnvService AgentEnvService) error {
	if agentEnvService != nil {
		vm.agentEnvService = agentEnvService
	}
	return nil
}

func (vm *softLayerHardware) SetVcapPassword(encryptedPwd string) (err error) {
	command := fmt.Sprintf("usermod -p '%s' vcap", encryptedPwd)
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to usermod vcap")
	}
	return
}

func (vm *softLayerHardware) Delete(agentID string) error {
	updateStateResponse, err := vm.baremetalClient.UpdateState(strconv.Itoa(vm.ID()), "bm.state.deleted")
	if err != nil || updateStateResponse.Status != 200 {
		return bosherr.WrapError(err, "Faled to call bms to delete baremetal")
	}

	command := "rm -f /var/vcap/bosh/*.json ; sv stop agent"
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	return err
}

func (vm *softLayerHardware) Reboot() error {
	return api.NotSupportedError{}
}

func (vm *softLayerHardware) ReloadOS(stemcell bslcstem.Stemcell) error {
	return api.NotSupportedError{}
}

func (vm *softLayerHardware) ReloadOSForBaremetal(stemcell string, netbootImage string) error {
	updateStateResponse, err := vm.baremetalClient.UpdateState(strconv.Itoa(vm.ID()), "bm.state.new")
	if err != nil || updateStateResponse.Status != 200 {
		return bosherr.WrapError(err, "Failed to call bms to update state of baremetal")
	}

	hardwareId, err := vm.provisionBaremetal(strconv.Itoa(vm.ID()), stemcell, netbootImage)
	if err != nil {
		return bosherr.WrapError(err, "Provision baremetal error")
	}

	if hardwareId == vm.ID() {
		return nil
	}

	return bosherr.Errorf("Failed to do os_reload against baremetal with id: %d", vm.ID())
}

func (vm *softLayerHardware) SetMetadata(vmMetadata VMMetadata) error {
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "set_vm_metadata not support for baremetal")
	return nil
}

func (vm *softLayerHardware) ConfigureNetworks(networks Networks) error {
	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from hardware with id: %d.", vm.ID())
	}

	oldAgentEnv.Networks = networks
	err = vm.agentEnvService.Update(oldAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring network setting on hardware with id: `%d`", vm.ID()))
	}

	return nil
}

func (vm *softLayerHardware) ConfigureNetworks2(networks Networks) error {
	return api.NotSupportedError{}
}

func (vm *softLayerHardware) AttachDisk(disk bslcdisk.Disk) error {
	volume, err := vm.fetchIscsiVolume(disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d`", disk.ID()))
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedHardware(disk.ID(), vm.ID())

	totalTime := time.Duration(0)
	if err == nil && allowed == false {
		for totalTime < slh.TIMEOUT {
			allowable, err := networkStorageService.AttachNetworkStorageToHardware(vm.hardware, disk.ID())
			if err != nil {
				if !strings.Contains(err.Error(), "HTTP error code") {
					return bosherr.WrapError(err, fmt.Sprintf("Granting volume access to virtual guest %d", vm.ID()))
				}
			} else {
				if allowable {
					break
				}
			}

			totalTime += slh.POLLING_INTERVAL
			time.Sleep(slh.POLLING_INTERVAL)
		}
	}
	if totalTime >= slh.TIMEOUT {
		return bosherr.Error("Waiting for grantting access to hardware TIME OUT!")
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript()
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from hardware `%d`", vm.ID()))
	}

	deviceName, err := vm.waitForVolumeAttached(volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to attach volume `%d` to hardware `%d`", disk.ID(), vm.ID()))
	}
	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from hardware with id: %d.", vm.ID())
	}

	var newAgentEnv AgentEnv
	if hasMultiPath {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/mapper/"+deviceName)
	} else {
		newAgentEnv = oldAgentEnv.AttachPersistentDisk(strconv.Itoa(disk.ID()), "/dev/"+deviceName)
	}

	err = vm.agentEnvService.Update(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on hardware with id: `%d`", vm.ID()))
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
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from hardware `%d`", vm.ID()))
	}

	err = vm.detachVolumeBasedOnShellScript(hasMultiPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to detach volume with id %d from hardware with id: %d.", volume.Id, vm.ID())
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedHardware(disk.ID(), vm.ID())
	if err == nil && allowed == true {
		err = networkStorageService.DetachNetworkStorageFromHardware(vm.hardware, disk.ID())
	}
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to revoke access of disk `%d` from hardware `%d`", disk.ID(), vm.ID()))
	}

	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from hardware with id: %d.", vm.ID())
	}

	newAgentEnv := oldAgentEnv.DetachPersistentDisk(strconv.Itoa(disk.ID()))
	err = vm.agentEnvService.Update(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on hardware with id: `%d`", vm.ID()))
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
				return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and hardware `%d`", disk.ID(), vm.ID()))
			}

			_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(volume)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to hardware `%d`", key, vm.ID()))
			}

			command := fmt.Sprintf("sleep 5; mount %s-part1 /var/vcap/store", devicePath)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
			if err != nil {
				return bosherr.WrapError(err, "mount /var/vcap/store")
			}
		}
	}

	return nil
}

func (vm *softLayerHardware) UpdateAgentEnv(agentEnv AgentEnv) error {
	return vm.agentEnvService.Update(agentEnv)
}

// Private methods
func (vm *softLayerHardware) waitForVolumeAttached(volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) (string, error) {

	oldDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(hasMultiPath)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from hardware `%d`", vm.ID()))
	}
	if len(oldDisks) > 2 {
		return "", bosherr.Error(fmt.Sprintf("Too manay persistent disks attached to hardware `%d`", vm.ID()))
	}

	credential, err := vm.getAllowedHostCredential()
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get iscsi host auth from hardware `%d`", vm.ID()))
	}

	_, err = vm.backupOpenIscsiConfBasedOnShellScript()
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to backup open iscsi conf files from hardware `%d`", vm.ID()))
	}

	_, err = vm.writeOpenIscsiInitiatornameBasedOnShellScript(credential)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi initiatorname from hardware `%d`", vm.ID()))
	}

	_, err = vm.writeOpenIscsiConfBasedOnShellScript(volume, credential)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi conf from hardware `%d`", vm.ID()))
	}

	_, err = vm.restartOpenIscsiBasedOnShellScript()
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to restart open iscsi from hardware `%d`", vm.ID()))
	}

	_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(volume)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to attach volume with id %d to hardware with id: %d.", volume.Id, vm.ID())
	}

	var deviceName string
	totalTime := time.Duration(0)
	for totalTime < slh.TIMEOUT {
		newDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(hasMultiPath)
		if err != nil {
			return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from hardware `%d`", vm.ID()))
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

		totalTime += slh.POLLING_INTERVAL
		time.Sleep(slh.POLLING_INTERVAL)
	}

	return "", bosherr.Errorf("Failed to attach disk '%d' to hardware '%d'", volume.Id, vm.ID())
}

func (vm *softLayerHardware) hasMulitPathToolBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("echo `command -v multipath`")
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
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
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command1)
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
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command2)
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

	volume, err := networkStorageService.GetNetworkStorage(volumeId)
	if err != nil {
		return datatypes.SoftLayer_Network_Storage{}, bosherr.WrapErrorf(err, "Cannot get iSCSI volume with id: %d", volumeId)
	}

	return volume, nil
}

func (vm *softLayerHardware) getAllowedHostCredential() (AllowedHostCredential, error) {
	hardwareService, err := vm.softLayerClient.GetSoftLayer_Hardware_Service()
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapError(err, "Cannot get softlayer hardware service.")
	}

	allowedHost, err := hardwareService.GetAllowedHost(vm.ID())
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapErrorf(err, "Cannot get allowed host with instance id: %d", vm.ID())
	}

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
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vm *softLayerHardware) restartOpenIscsiBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vm *softLayerHardware) discoveryOpenIscsiTargetsBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage) (bool, error) {
	command := fmt.Sprintf("sleep 5; iscsiadm -m discovery -t sendtargets -p %s", volume.ServiceResourceBackendIpAddress)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "discoverying open iscsi targets")
	}

	command = "sleep 5; echo `iscsiadm -m node -l`"
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "login iscsi targets")
	}

	return true, nil
}

func (vm *softLayerHardware) writeOpenIscsiInitiatornameBasedOnShellScript(credential AllowedHostCredential) (bool, error) {
	if len(credential.Iqn) > 0 {
		command := fmt.Sprintf("echo 'InitiatorName=%s' > /etc/iscsi/initiatorname.iscsi", credential.Iqn)
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
		if err != nil {
			return false, bosherr.WrapError(err, "Writing to /etc/iscsi/initiatorname.iscsi")
		}
	}

	return true, nil
}

func (vm *softLayerHardware) writeOpenIscsiConfBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage, credential AllowedHostCredential) (bool, error) {
	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("open_iscsid_conf").Parse(EtcIscsidConfTemplate))
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

func (vm *softLayerHardware) detachVolumeBasedOnShellScript(hasMultiPath bool) error {
	// umount /var/vcap/store in case read-only mount
	isMounted, err := vm.isMountPoint("/var/vcap/store")
	if err != nil {
		return bosherr.WrapError(err, "check mount point /var/vcap/store")
	}

	if isMounted {
		step00 := fmt.Sprintf("umount -l /var/vcap/store")
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step00)
		if err != nil {
			return bosherr.WrapError(err, "umount -l /var/vcap/store")
		}
		vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "umount -l /var/vcap/store", nil)
	}

	// stop open-iscsi
	step1 := fmt.Sprintf("/etc/init.d/open-iscsi stop")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step1)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "/etc/init.d/open-iscsi stop", nil)

	// clean up /etc/iscsi/send_targets/
	step2 := fmt.Sprintf("rm -rf /etc/iscsi/send_targets")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step2)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets")
	}
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "rm -rf /etc/iscsi/send_targets", nil)

	// clean up /etc/iscsi/nodes/
	step3 := fmt.Sprintf("rm -rf /etc/iscsi/nodes")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step3)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes")
	}

	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "rm -rf /etc/iscsi/nodes", nil)

	// start open-iscsi
	step4 := fmt.Sprintf("/etc/init.d/open-iscsi start")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step4)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "/etc/init.d/open-iscsi start", nil)

	if hasMultiPath {
		// restart dm-multipath tool
		step5 := fmt.Sprintf("service multipath-tools restart")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step5)
		if err != nil {
			return bosherr.WrapError(err, "Restarting Multipath deamon")
		}
		vm.logger.Debug(SOFTLAYER_HARDWARE_LOG_TAG, "service multipath-tools restart", nil)
	}

	return nil
}

func (vm *softLayerHardware) isMountPoint(path string) (bool, error) {
	mounts, err := vm.searchMounts()
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

func (vm *softLayerHardware) searchMounts() ([]Mount, error) {
	var mounts []Mount
	stdout, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), "mount")
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

func (vm *softLayerHardware) provisionBaremetal(server_id string, stemcell string, netboot_image string) (int, error) {
	provisioningBaremetalInfo := bmscl.ProvisioningBaremetalInfo{
		VmNamePrefix:     server_id,
		Bm_stemcell:      stemcell,
		Bm_netboot_image: netboot_image,
	}
	createBaremetalResponse, err := vm.baremetalClient.ProvisioningBaremetal(provisioningBaremetalInfo)
	if err != nil || createBaremetalResponse.Status != 200 || createBaremetalResponse.Data.TaskId == 0 {
		return 0, bosherr.WrapErrorf(err, "Failed to provisioning baremetal")
	}

	task_id := createBaremetalResponse.Data.TaskId
	slh.TIMEOUT = 10 * time.Minute
	totalTime := time.Duration(0)
	for totalTime < slh.TIMEOUT {
		taskOutput, err := vm.baremetalClient.TaskJsonOutput(task_id, "task")
		if err != nil {
			return 0, bosherr.WrapErrorf(err, "Failed to get state with task_id: %d", task_id)
		}

		info := taskOutput.Data["info"].(map[string]interface{})
		switch info["status"].(string) {
		case "failed":
			return 0, bosherr.Errorf("Failed to install the stemcell: %v", taskOutput)

		case "completed":
			serverOutput, err := vm.baremetalClient.TaskJsonOutput(task_id, "server")
			if err != nil {
				return 0, bosherr.WrapErrorf(err, "Failed to get server_id with task_id: %d", task_id)
			}
			info = serverOutput.Data["info"].(map[string]interface{})
			return int(info["id"].(float64)), nil
		default:
			totalTime += slh.POLLING_INTERVAL
			time.Sleep(slh.POLLING_INTERVAL)
		}
	}

	return 0, bosherr.Error("Provisioning baremetal timeout")
}
