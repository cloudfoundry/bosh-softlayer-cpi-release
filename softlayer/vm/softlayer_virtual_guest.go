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

	slh "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common/helper"
	sl "github.com/maximilien/softlayer-go/softlayer"

	. "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"

	api "github.com/cloudfoundry/bosh-softlayer-cpi/api"
	"github.com/cloudfoundry/bosh-softlayer-cpi/util"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

type softLayerVirtualGuest struct {
	id int

	virtualGuest datatypes.SoftLayer_Virtual_Guest

	softLayerClient sl.Client
	sshClient       util.SshClient

	agentEnvService AgentEnvService

	logger boshlog.Logger
}

func NewSoftLayerVirtualGuest(virtualGuest datatypes.SoftLayer_Virtual_Guest, softLayerClient sl.Client, sshClient util.SshClient, logger boshlog.Logger) VM {
	slh.TIMEOUT = 60 * time.Minute
	slh.POLLING_INTERVAL = 10 * time.Second

	return &softLayerVirtualGuest{
		id: virtualGuest.Id,

		virtualGuest: virtualGuest,

		softLayerClient: softLayerClient,
		sshClient:       sshClient,

		logger: logger,
	}
}

func (vm *softLayerVirtualGuest) ID() int { return vm.id }

func (vm *softLayerVirtualGuest) GetDataCenterId() int {
	return vm.virtualGuest.Datacenter.Id
}

func (vm *softLayerVirtualGuest) GetPrimaryIP() string {
	return vm.virtualGuest.PrimaryIpAddress
}

func (vm *softLayerVirtualGuest) GetPrimaryBackendIP() string {
	return vm.virtualGuest.PrimaryBackendIpAddress
}

func (vm *softLayerVirtualGuest) GetRootPassword() string {
	passwords := vm.virtualGuest.OperatingSystem.Passwords
	for _, password := range passwords {
		if password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}

	return ""
}

func (vm *softLayerVirtualGuest) GetFullyQualifiedDomainName() string {
	return vm.virtualGuest.FullyQualifiedDomainName
}

func (vm *softLayerVirtualGuest) Delete(agentID string) error {
	return nil
}

func (vm *softLayerVirtualGuest) SetVcapPassword(encryptedPwd string) (err error) {
	command := fmt.Sprintf("usermod -p '%s' vcap", encryptedPwd)
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to usermod vcap")
	}
	return
}

func (vm *softLayerVirtualGuest) SetAgentEnvService(agentEnvService AgentEnvService) error {
	if agentEnvService != nil {
		vm.agentEnvService = agentEnvService
	}
	return nil
}

func (vm *softLayerVirtualGuest) Reboot() error {
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

func (vm *softLayerVirtualGuest) ReloadOS(stemcell bslcstem.Stemcell) error {
	reload_OS_Config := sldatatypes.Image_Template_Config{
		ImageTemplateId: strconv.Itoa(stemcell.ID()),
	}

	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	err = slh.WaitForVirtualGuestToHaveNoRunningTransactions(vm.softLayerClient, vm.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for VirtualGuest %d to have no pending transactions before os reload", vm.ID()))
	}
	vm.logger.Info(SOFTLAYER_VM_OS_RELOAD_TAG, fmt.Sprintf("No transaction is running on this VM %d", vm.ID()))

	err = virtualGuestService.ReloadOperatingSystem(vm.ID(), reload_OS_Config)
	if err != nil {
		return bosherr.WrapError(err, "Failed to reload OS on the specified VirtualGuest from SoftLayer client")
	}

	err = vm.postCheckActiveTransactionsForOSReload(vm.softLayerClient)
	if err != nil {
		return err
	}

	return nil
}

func (vm *softLayerVirtualGuest) ReloadOSForBaremetal(string, string) error {
	return api.NotSupportedError{}
}

func (vm *softLayerVirtualGuest) SetMetadata(vmMetadata VMMetadata) error {
	tags, err := vm.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return err
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

func (vm *softLayerVirtualGuest) ConfigureNetworks(networks Networks) error {
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

type sshClientWrapper struct {
	client   util.SshClient
	ip       string
	user     string
	password string
}

func (s *sshClientWrapper) Output(command string) ([]byte, error) {
	o, err := s.client.ExecCommand(s.user, s.password, s.ip, command)
	return []byte(o), err
}

func (vm *softLayerVirtualGuest) ConfigureNetworks2(networks Networks) error {
	ubuntu := Ubuntu{
		SoftLayerClient: vm.softLayerClient.GetHttpClient(),
		SSHClient: &sshClientWrapper{
			client:   vm.sshClient,
			ip:       vm.GetPrimaryBackendIP(),
			user:     ROOT_USER_NAME,
			password: vm.GetRootPassword(),
		},
		SoftLayerFileService: NewSoftlayerFileService(util.GetSshClient(), vm.logger),
	}

	err := ubuntu.ConfigureNetwork(networks, vm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to configure networking for virtual guest with id: %d.", vm.ID())
	}

	return nil
}

func (vm *softLayerVirtualGuest) AttachDisk(disk bslcdisk.Disk) error {
	volume, err := vm.fetchIscsiVolume(disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d`", disk.ID()))
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())

	totalTime := time.Duration(0)
	if err == nil && allowed == false {
		for totalTime < slh.TIMEOUT {
			allowable, err := networkStorageService.AttachNetworkStorageToVirtualGuest(vm.virtualGuest, disk.ID())
			if err != nil {
				if !strings.Contains(err.Error(), "A Volume Provisioning is currently in progress on volume") {
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

func (vm *softLayerVirtualGuest) DetachDisk(disk bslcdisk.Disk) error {
	volume, err := vm.fetchIscsiVolume(disk.ID())
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("failed in disk `%d`", disk.ID()))
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript()
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to get multipath information from virtual guest `%d`", vm.ID()))
	}

	err = vm.detachVolumeBasedOnShellScript(volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to detach volume with id %d from virtual guest with id: %d.", volume.Id, vm.ID())
	}

	networkStorageService, err := vm.softLayerClient.GetSoftLayer_Network_Storage_Service()
	if err != nil {
		return bosherr.WrapError(err, "Cannot get network storage service.")
	}

	allowed, err := networkStorageService.HasAllowedVirtualGuest(disk.ID(), vm.ID())
	if err == nil && allowed == true {
		err = networkStorageService.DetachNetworkStorageFromVirtualGuest(vm.virtualGuest, disk.ID())
	}
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to revoke access of disk `%d` from virtual gusest `%d`", disk.ID(), vm.ID()))
	}

	oldAgentEnv, err := vm.agentEnvService.Fetch()
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to unmarshal userdata from virutal guest with id: %d.", vm.ID())
	}

	newAgentEnv := oldAgentEnv.DetachPersistentDisk(strconv.Itoa(disk.ID()))
	err = vm.UpdateAgentEnv(newAgentEnv)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Configuring userdata on VirtualGuest with id: `%d`", vm.ID()))
	}

	if len(newAgentEnv.Disks.Persistent) == 1 {
		for key, devicePath := range newAgentEnv.Disks.Persistent {
			leftDiskId, err := strconv.Atoi(key)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to transfer disk id %s from string to int", key))
			}
			vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Left Disk Id %d", leftDiskId)
			vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Left Disk device path %s", devicePath)
			volume, err := vm.fetchIscsiVolume(leftDiskId)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), vm.ID()))
			}

			_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(volume)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to virtual guest `%d`", key, vm.ID()))
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

func (vm *softLayerVirtualGuest) UpdateAgentEnv(agentEnv AgentEnv) error {
	return vm.agentEnvService.Update(agentEnv)
}

// Private methods
func (vm *softLayerVirtualGuest) extractTagsFromVMMetadata(vmMetadata VMMetadata) ([]string, error) {
	tags := []string{}
	status := ""
	for key, value := range vmMetadata {
		if key == "compiling" || key == "job" || key == "index" || key == "deployment" || key == "deleted" {
			stringValue, err := value.(string)
			if !err {
				return []string{}, bosherr.Errorf("Cannot convert tags metadata value `%v` to string", value)
			}

			if status == "" {
				status = key + ":" + stringValue
			} else {
				status = status + "," + key + ":" + stringValue
			}
		}
		tags = vm.parseTags(status)
	}

	return tags, nil
}

func (vm *softLayerVirtualGuest) parseTags(value string) []string {
	return strings.Split(value, ",")
}

func (vm *softLayerVirtualGuest) waitForVolumeAttached(volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) (string, error) {

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
	for totalTime < slh.TIMEOUT {
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

		totalTime += slh.POLLING_INTERVAL
		time.Sleep(slh.POLLING_INTERVAL)
	}

	return "", bosherr.Errorf("Failed to attach disk '%d' to virtual guest '%d'", volume.Id, vm.ID())
}

func (vm *softLayerVirtualGuest) hasMulitPathToolBasedOnShellScript() (bool, error) {
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

func (vm *softLayerVirtualGuest) getIscsiDeviceNamesBasedOnShellScript(hasMultiPath bool) ([]string, error) {
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
		vm.logger.Info(SOFTLAYER_VM_LOG_TAG, fmt.Sprintf("Devices on VM %d: %s", vm.ID(), result))
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

		vm.logger.Info(SOFTLAYER_VM_LOG_TAG, fmt.Sprintf("Devices on VM %d: %s", vm.ID(), result))
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

func (vm *softLayerVirtualGuest) fetchIscsiVolume(volumeId int) (datatypes.SoftLayer_Network_Storage, error) {
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

func (vm *softLayerVirtualGuest) getAllowedHostCredential() (AllowedHostCredential, error) {
	var allowedHost datatypes.SoftLayer_Network_Storage_Allowed_Host
	var err error

	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return AllowedHostCredential{}, bosherr.WrapError(err, "Cannot get softlayer virtual guest service.")
	}

	allowedHost, err = virtualGuestService.GetAllowedHost(vm.ID())
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

func (vm *softLayerVirtualGuest) backupOpenIscsiConfBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("cp /etc/iscsi/iscsid.conf{,.save}")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) restartOpenIscsiBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) discoveryOpenIscsiTargetsBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage) (bool, error) {
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

func (vm *softLayerVirtualGuest) writeOpenIscsiInitiatornameBasedOnShellScript(credential AllowedHostCredential) (bool, error) {
	if len(credential.Iqn) > 0 {
		command := fmt.Sprintf("echo 'InitiatorName=%s' > /etc/iscsi/initiatorname.iscsi", credential.Iqn)
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), command)
		if err != nil {
			return false, bosherr.WrapError(err, "Writing to /etc/iscsi/initiatorname.iscsi")
		}
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) writeOpenIscsiConfBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage, credential AllowedHostCredential) (bool, error) {
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

func (vm *softLayerVirtualGuest) detachVolumeBasedOnShellScript(volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) error {
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
		vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "umount -l /var/vcap/store", nil)
	}

	// stop open-iscsi
	step1 := fmt.Sprintf("/etc/init.d/open-iscsi stop")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step1)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "/etc/init.d/open-iscsi stop", nil)

	// clean up /etc/iscsi/send_targets/
	step2 := fmt.Sprintf("rm -rf /etc/iscsi/send_targets")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step2)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "rm -rf /etc/iscsi/send_targets", nil)

	// clean up /etc/iscsi/nodes/
	step3 := fmt.Sprintf("rm -rf /etc/iscsi/nodes")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step3)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes")
	}

	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "rm -rf /etc/iscsi/nodes", nil)

	// start open-iscsi
	step4 := fmt.Sprintf("/etc/init.d/open-iscsi start")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step4)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "/etc/init.d/open-iscsi start", nil)

	if hasMultiPath {
		// restart dm-multipath tool
		step5 := fmt.Sprintf("service multipath-tools restart")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.GetRootPassword(), vm.GetPrimaryBackendIP(), step5)
		if err != nil {
			return bosherr.WrapError(err, "Restarting Multipath deamon")
		}
		vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "service multipath-tools restart", nil)
	}

	return nil
}

func (vm *softLayerVirtualGuest) postCheckActiveTransactionsForOSReload(softLayerClient sl.Client) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < slh.TIMEOUT {
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

		totalTime += slh.POLLING_INTERVAL
		time.Sleep(slh.POLLING_INTERVAL)
	}

	if totalTime >= slh.TIMEOUT {
		return errors.New(fmt.Sprintf("Waiting for OS Reload transaction to start TIME OUT!"))
	}

	err = slh.WaitForVirtualGuest(vm.softLayerClient, vm.ID(), "RUNNING")
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP error code") {
			return bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id %d", vm.ID()))
		}
	}

	vm.logger.Info(SOFTLAYER_VM_OS_RELOAD_TAG, fmt.Sprintf("The virtual guest %d is powered on", vm.ID()))

	return nil
}

func (vm *softLayerVirtualGuest) postCheckActiveTransactionsForDeleteVM(softLayerClient sl.Client, virtualGuestId int) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < slh.TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(virtualGuestId)
		if err != nil {
			if !strings.Contains(err.Error(), "HTTP error code") {
				return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
			}
		}

		if len(activeTransactions) > 0 {
			vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "Delete VM transaction started", nil)
			break
		}

		totalTime += slh.POLLING_INTERVAL
		time.Sleep(slh.POLLING_INTERVAL)
	}

	if totalTime >= slh.TIMEOUT {
		return errors.New(fmt.Sprintf("Waiting for DeleteVM transaction to start TIME OUT!"))
	}

	totalTime = time.Duration(0)
	for totalTime < slh.TIMEOUT {
		vm1, err := virtualGuestService.GetObject(virtualGuestId)
		if err != nil || vm1.Id == 0 {
			vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "VM doesn't exist. Delete done", nil)
			break
		}

		activeTransaction, err := virtualGuestService.GetActiveTransaction(virtualGuestId)
		if err != nil {
			if !strings.Contains(err.Error(), "HTTP error code") {
				return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
			}
		}

		averageDuration := activeTransaction.TransactionStatus.AverageDuration
		if strings.HasPrefix(averageDuration, ".") || averageDuration == "" {
			averageDuration = "0" + averageDuration
		}

		averageTransactionDuration, err := strconv.ParseFloat(averageDuration, 32)
		if err != nil {
			averageTransactionDuration = 0
		}

		if averageTransactionDuration > 30 {
			vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "Deleting VM instance had been launched and it is a long transaction. Please check Softlayer Portal", nil)
			break
		}

		vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "This is a short transaction, waiting for all active transactions to complete", nil)
		totalTime += slh.POLLING_INTERVAL
		time.Sleep(slh.POLLING_INTERVAL)
	}

	if totalTime >= slh.TIMEOUT {
		return errors.New(fmt.Sprintf("After deleting a vm, waiting for active transactions to complete TIME OUT!"))
	}

	return nil
}

func (vm *softLayerVirtualGuest) isMountPoint(path string) (bool, error) {
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

func (vm *softLayerVirtualGuest) searchMounts() ([]Mount, error) {
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
