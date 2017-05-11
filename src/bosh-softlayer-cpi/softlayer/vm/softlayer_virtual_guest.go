package vm

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

	bsl "bosh-softlayer-cpi/softlayer/client"

	. "bosh-softlayer-cpi/softlayer/common"
	bslcdisk "bosh-softlayer-cpi/softlayer/disk"
	bslcstem "bosh-softlayer-cpi/softlayer/stemcell"

	bslnet "bosh-softlayer-cpi/softlayer/networks"
	"bosh-softlayer-cpi/util"
	datatypes "github.com/softlayer/softlayer-go/datatypes"
)

type softLayerVirtualGuest struct {
	virtualGuest *datatypes.Virtual_Guest

	softLayerClient bsl.Client
	sshClient       util.SshClient

	agentEnvService AgentEnvService

	logger boshlog.Logger
}

func NewSoftLayerVirtualGuest(virtualGuest *datatypes.Virtual_Guest, softLayerClient bsl.Client, sshClient util.SshClient, logger boshlog.Logger) VM {
	return &softLayerVirtualGuest{
		virtualGuest: virtualGuest,

		softLayerClient: softLayerClient,
		sshClient:       sshClient,

		logger: logger,
	}
}

func (vm *softLayerVirtualGuest) ID() *int { return vm.virtualGuest.Id }

func (vm *softLayerVirtualGuest) GetDataCenter() *string {
	return vm.virtualGuest.Datacenter.Name
}

func (vm *softLayerVirtualGuest) GetPrimaryIP() *string {
	return vm.virtualGuest.PrimaryIpAddress
}

func (vm *softLayerVirtualGuest) GetPrimaryBackendIP() *string {
	return vm.virtualGuest.PrimaryBackendIpAddress
}

func (vm *softLayerVirtualGuest) GetRootPassword() *string {
	passwords := vm.virtualGuest.OperatingSystem.Passwords
	for _, password := range passwords {
		if *password.Username == ROOT_USER_NAME {
			return password.Password
		}
	}

	return nil
}

func (vm *softLayerVirtualGuest) GetFullyQualifiedDomainName() *string {
	return vm.virtualGuest.FullyQualifiedDomainName
}

func (vm *softLayerVirtualGuest) SetVcapPassword(encryptedPwd string) (err error) {
	command := fmt.Sprintf("usermod -p '%s' vcap", encryptedPwd)
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
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
	err := vm.softLayerClient.RebootInstance(*vm.ID(), true, false)
	if err != nil {
		return bosherr.WrapError(err, "Rebooting (soft) SoftLayer VirtualGuest from client")
	}

	return nil
}

func (vm *softLayerVirtualGuest) ReloadOS(stemcell bslcstem.Stemcell) error {
	return vm.softLayerClient.ReloadInstance(*vm.ID(), stemcell.ID())
}

func (vm *softLayerVirtualGuest) SetMetadata(vmMetadata VMMetadata) error {
	tags, err := vm.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return bosherr.WrapError(err, "Extracting tags from vm metadata")
	}

	err = vm.softLayerClient.SetTags(*vm.ID(), tags)
	if err != nil {
		return bosherr.WrapErrorf(err, "Settings tags on virtualGuest `%d`", vm.ID())
	}

	return nil
}

func (vm *softLayerVirtualGuest) ConfigureNetworksSettings(networks bslnet.Networks) error {
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

func (vm *softLayerVirtualGuest) ConfigureNetworks(networks bslnet.Networks) (bslnet.Networks, error) {
	vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "Configuring networks: %#v", networks)
	ubuntu := bslnet.Softlayer_Ubuntu_Net{
		LinkNamer: bslnet.NewIndexedNamer(networks),
	}

	componentByNetwork, err := ubuntu.ComponentByNetworkName(*vm.virtualGuest, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Mapping network component and name")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "ComponentByNetworkName: %#v", componentByNetwork)

	networks, err = ubuntu.NormalizeNetworkDefinitions(networks, componentByNetwork)
	if err != nil {
		return networks, bosherr.WrapError(err, "Normalizing network definitions")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Normalized networks: %#v", networks)

	networks, err = ubuntu.NormalizeDynamics(*vm.virtualGuest, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Normalizing dynamic networks definitions")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Normalized Dynamics: %#v", networks)

	componentByNetwork, err = ubuntu.ComponentByNetworkName(*vm.virtualGuest, networks)
	if err != nil {
		return networks, bosherr.WrapError(err, "Mapping network component and name")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "ComponentByNetworkName: %#v", componentByNetwork)

	networks, err = ubuntu.FinalizedNetworkDefinitions(*vm.virtualGuest, networks, componentByNetwork)
	if err != nil {
		return networks, bosherr.WrapError(err, "Finalizing networks definitions")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Finalized network definition: %#v", networks)

	return networks, nil
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

func (vm *softLayerVirtualGuest) AttachDisk(disk bslcdisk.Disk) error {
	volume, err := vm.softLayerClient.GetBlockVolumeDetails(disk.ID(), bsl.VOLUME_DETAIL_MASK)
	if err != nil {
		return bosherr.WrapErrorf(err, "Fetching volume details with id `%d`", disk.ID())
	}

	until := time.Now().Add(time.Duration(1) * time.Hour)
	err = vm.softLayerClient.AuthorizeHostToVolume(vm.virtualGuest, disk.ID(), until)
	if err != nil {
		return bosherr.WrapErrorf(err, "Authorizing vm with id `%d` to disk with id `%d`", vm.ID(), disk.ID())
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting multipath information from vm with id `%d`", vm.ID())
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
	volume, err := vm.softLayerClient.GetBlockVolumeDetails(disk.ID(), bsl.VOLUME_DETAIL_MASK)
	if err != nil {
		return bosherr.WrapErrorf(err, "Fetching volume details with id `%d`", disk.ID())
	}

	hasMultiPath, err := vm.hasMulitPathToolBasedOnShellScript()
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Getting multipath information from vm with id `%d`", vm.ID()))
	}

	err = vm.detachVolumeBasedOnShellScript(volume, hasMultiPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to detach volume with id %d from virtual guest with id: %d.", volume.Id, vm.ID())
	}

	until := time.Now().Add(time.Duration(1) * time.Hour)
	err = vm.softLayerClient.DeauthorizeHostToVolume(vm.virtualGuest, disk.ID(), until)
	if err != nil {
		return bosherr.WrapErrorf(err, "De-Authorizing vm with id `%d` to disk with id `%d`", vm.ID(), disk.ID())
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
			volume, err := vm.softLayerClient.GetBlockVolumeDetails(leftDiskId, bsl.VOLUME_DETAIL_MASK)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), vm.ID()))
			}

			_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(volume)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to virtual guest `%d`", key, vm.ID()))
			}

			command := fmt.Sprintf("sleep 5; mount %s-part1 /var/vcap/store", devicePath)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
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

func (vm *softLayerVirtualGuest) DeleteAgentEnv() error {
	return vm.agentEnvService.Delete()
}

// Private methods
func (vm *softLayerVirtualGuest) extractTagsFromVMMetadata(vmMetadata VMMetadata) (string, error) {
	var tagStringBuffer bytes.Buffer
	var i int
	for key, value := range vmMetadata {
		if key == "compiling" || key == "job" || key == "index" || key == "deployment" || key == "deleted" {
			stringValue, err := value.(string)
			if !err {
				return "", bosherr.Errorf("Converting tags metadata value `%v` to string failed", value)
			}
			tagStringBuffer.WriteString(key + ":" + stringValue)
			if i != len(vmMetadata)-1 {
				tagStringBuffer.WriteString(", ")
			}
		}
		i++
	}

	return tagStringBuffer.String(), nil
}

func (vm *softLayerVirtualGuest) waitForVolumeAttached(volume datatypes.Network_Storage, hasMultiPath bool) (string, error) {
	oldDisks, err := vm.getIscsiDeviceNamesBasedOnShellScript(hasMultiPath)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to get devices names from virtual guest `%d`", vm.ID()))
	}
	if len(oldDisks) > 2 {
		return "", bosherr.Error(fmt.Sprintf("Too manay persistent disks attached to virtual guest `%d`", vm.ID()))
	}

	credential, err := vm.softLayerClient.GetAllowedHostCredential(*vm.ID())
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
	for totalTime < 5*time.Minute {
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

		totalTime += 5 * time.Second
		time.Sleep(5 * time.Second)
	}

	return "", bosherr.Errorf("Failed to attach disk '%d' to virtual guest '%d'", volume.Id, vm.ID())
}

func (vm *softLayerVirtualGuest) hasMulitPathToolBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("echo `command -v multipath`")
	output, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
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
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command1)
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
		result, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command2)
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

func (vm *softLayerVirtualGuest) backupOpenIscsiConfBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("cp /etc/iscsi/iscsid.conf{,.save}")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) restartOpenIscsiBasedOnShellScript() (bool, error) {
	command := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) discoveryOpenIscsiTargetsBasedOnShellScript(volume datatypes.Network_Storage) (bool, error) {
	command := fmt.Sprintf("sleep 5; iscsiadm -m discovery -t sendtargets -p %s", volume.ServiceResourceBackendIpAddress)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "discoverying open iscsi targets")
	}

	command = "sleep 5; echo `iscsiadm -m node -l`"
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
	if err != nil {
		return false, bosherr.WrapError(err, "login iscsi targets")
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) writeOpenIscsiInitiatornameBasedOnShellScript(credential datatypes.Network_Storage_Allowed_Host) (bool, error) {
	if len(*credential.Name) > 0 {
		command := fmt.Sprintf("echo 'InitiatorName=%s' > /etc/iscsi/initiatorname.iscsi", credential.Name)
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), command)
		if err != nil {
			return false, bosherr.WrapError(err, "Writing to /etc/iscsi/initiatorname.iscsi")
		}
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) writeOpenIscsiConfBasedOnShellScript(volume datatypes.Network_Storage, allowedHost datatypes.Network_Storage_Allowed_Host) (bool, error) {
	type credential struct {
		Username string
		Password string
	}

	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("open_iscsid_conf").Parse(EtcIscsidConfTemplate))
	err := t.Execute(buffer, credential{
		Username: *allowedHost.Credential.Username,
		Password: *allowedHost.Credential.Password,
	})
	if err != nil {
		return false, bosherr.WrapError(err, "Generating config from template")
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

	if err = vm.sshClient.UploadFile(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), file.Name(), "/etc/iscsi/iscsid.conf"); err != nil {
		return false, bosherr.WrapError(err, "Writing to /etc/iscsi/iscsid.conf")
	}

	return true, nil
}

func (vm *softLayerVirtualGuest) detachVolumeBasedOnShellScript(volume datatypes.Network_Storage, hasMultiPath bool) error {
	// umount /var/vcap/store in case read-only mount
	isMounted, err := vm.isMountPoint("/var/vcap/store")
	if err != nil {
		return bosherr.WrapError(err, "check mount point /var/vcap/store")
	}

	if isMounted {
		step00 := fmt.Sprintf("umount -l /var/vcap/store")
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), step00)
		if err != nil {
			return bosherr.WrapError(err, "umount -l /var/vcap/store")
		}
		vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "umount -l /var/vcap/store", nil)
	}

	// stop open-iscsi
	step1 := fmt.Sprintf("/etc/init.d/open-iscsi stop")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), step1)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "/etc/init.d/open-iscsi stop", nil)

	// clean up /etc/iscsi/send_targets/
	step2 := fmt.Sprintf("rm -rf /etc/iscsi/send_targets")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), step2)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "rm -rf /etc/iscsi/send_targets", nil)

	// clean up /etc/iscsi/nodes/
	step3 := fmt.Sprintf("rm -rf /etc/iscsi/nodes")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), step3)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes")
	}

	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "rm -rf /etc/iscsi/nodes", nil)

	// start open-iscsi
	step4 := fmt.Sprintf("/etc/init.d/open-iscsi start")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), step4)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "/etc/init.d/open-iscsi start", nil)

	if hasMultiPath {
		// restart dm-multipath tool
		step5 := fmt.Sprintf("service multipath-tools restart")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), step5)
		if err != nil {
			return bosherr.WrapError(err, "Restarting Multipath deamon")
		}
		vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "service multipath-tools restart", nil)
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
	stdout, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, *vm.GetRootPassword(), *vm.GetPrimaryBackendIP(), "mount")
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
