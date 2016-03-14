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

	bslcommon "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/common"
	bslcdisk "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvmpool "github.com/cloudfoundry/bosh-softlayer-cpi/softlayer/vm/pool"

	common "github.com/cloudfoundry/bosh-softlayer-cpi/common"
	util "github.com/cloudfoundry/bosh-softlayer-cpi/util"
	datatypes "github.com/maximilien/softlayer-go/data_types"
	sldatatypes "github.com/maximilien/softlayer-go/data_types"
)

const (
	SOFTLAYER_VM_OS_RELOAD_TAG = "OSReload"
	SOFTLAYER_VM_LOG_TAG       = "SoftLayerVM"
	ROOT_USER_NAME             = "root"
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

func (vm SoftLayerVM) Delete(agentID string) error {
	if strings.ToUpper(common.GetOSEnvVariable("OS_RELOAD_ENABLED", "TRUE")) == "FALSE" {
		if strings.ToUpper(common.GetOSEnvVariable("DEL_NOT_ALLOWED", "FALSE")) == "FALSE" {
			return vm.DeleteVM()
		} else {
			return bosherr.Error("DEL_NOT_ALLOWED is set to TRUE, the VM deletion reqeust is refused.")
		}
	}

	err := bslcvmpool.InitVMPoolDB(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL, vm.logger)
	if err != nil {
		return bosherr.WrapError(err, "Failed to initialize VM pool DB")
	}

	db, err := bslcvmpool.OpenDB(bslcvmpool.SQLITE_DB_FILE_PATH)
	if err != nil {
		return bosherr.WrapError(err, "Opening DB")
	}

	vmInfoDB := bslcvmpool.NewVMInfoDB(vm.id, "", "", "", "", vm.logger, db)
	defer vmInfoDB.CloseDB()

	err = vmInfoDB.QueryVMInfobyID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Failed to query VM info by given ID %d", vm.id))
	}
	vm.logger.Info(SOFTLAYER_VM_LOG_TAG, fmt.Sprintf("vmInfoDB.vmProperties.id is %d", vmInfoDB.VmProperties.Id))

	if vmInfoDB.VmProperties.Id != 0 {
		if agentID != "" {
			vm.logger.Info(SOFTLAYER_VM_LOG_TAG, fmt.Sprintf("Release the VM with id %d back to the VM pool", vmInfoDB.VmProperties.Id))
			vmInfoDB.VmProperties.InUse = "f"
			err = vmInfoDB.UpdateVMInfoByID(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to update in_use to %s by given ID %d", vmInfoDB.VmProperties.InUse, vm.id))
			} else {
				return nil
			}
		} else {
			if strings.ToUpper(common.GetOSEnvVariable("DEL_NOT_ALLOWED", "FALSE")) == "FALSE" {
				return vm.DeleteVM()
			} else {
				return bosherr.Error("DEL_NOT_ALLOWED is set to TRUE, the VM deletion reqeust is refused.")
			}
		}
	} else {
		if strings.ToUpper(common.GetOSEnvVariable("DEL_NOT_ALLOWED", "FALSE")) == "FALSE" {
			return vm.DeleteVM()
		} else {
			return bosherr.Error("DEL_NOT_ALLOWED is set to TRUE, the VM deletion reqeust is refused.")
		}
	}

}

func (vm SoftLayerVM) DeleteVM() error {
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

	if strings.ToUpper(common.GetOSEnvVariable("OS_RELOAD_ENABLED", "TRUE")) == "TRUE" {
		db, err := bslcvmpool.OpenDB(bslcvmpool.SQLITE_DB_FILE_PATH)
		if err != nil {
			return bosherr.WrapError(err, "Opening DB")
		}

		vmInfoDB := bslcvmpool.NewVMInfoDB(vm.ID(), "", "", "", "", vm.logger, db)
		err = vmInfoDB.DeleteVMFromVMDB(bslcvmpool.DB_RETRY_TIMEOUT, bslcvmpool.DB_RETRY_INTERVAL)
		if err != nil {
			return bosherr.WrapError(err, "Failed to delete the record from VM pool DB")
		}
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

func (vm SoftLayerVM) ReloadOS(stemcell bslcstem.Stemcell) error {
	reload_OS_Config := sldatatypes.Image_Template_Config{
		ImageTemplateId: strconv.Itoa(stemcell.ID()),
	}

	virtualGuestService, err := vm.softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	err = bslcommon.WaitForVirtualGuestToHaveNoRunningTransactions(vm.softLayerClient, vm.ID())
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

func (vm SoftLayerVM) SetMetadata(vmMetadata VMMetadata) error {
	tags, err := vm.extractTagsFromVMMetadata(vmMetadata)
	if err != nil {
		return err
	}

	//Check below needed since Golang strings.Split return [""] on strings.Split("", ",")
	if len(tags) == 1 && tags[0] == "" {
		return nil
	}

	if len(tags) == 0 {
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
		for key, devicePath := range newAgentEnv.Disks.Persistent {
			leftDiskId, err := strconv.Atoi(key)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to transfer disk id %s from string to int", key))
			}
			vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Left Disk Id %d", leftDiskId)
			vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "Left Disk device path %s", devicePath)
			virtualGuest, volume, err := vm.fetchVMandIscsiVolume(vm.ID(), leftDiskId)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to fetch disk `%d` and virtual gusest `%d`", disk.ID(), virtualGuest.Id))
			}

			_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(virtualGuest, volume)
			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Failed to reattach volume `%s` to virtual guest `%d`", key, virtualGuest.Id))
			}

			command := fmt.Sprintf("sleep 5; mount %s-part1 /var/vcap/store", devicePath)
			_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
			if err != nil {
				return bosherr.WrapError(err, "mount /var/vcap/store")
			}
		}
	}

	return nil
}

// Private methods
func (vm SoftLayerVM) extractTagsFromVMMetadata(vmMetadata VMMetadata) ([]string, error) {
	tags := []string{}
	status := ""
	for key, value := range vmMetadata {
		if key == "compiling" || key == "job" || key == "index" || key == "deployment" {
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

func (vm SoftLayerVM) parseTags(value string) []string {
	return strings.Split(value, ",")
}

func (vm SoftLayerVM) waitForVolumeAttached(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) (string, error) {

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

	_, err = vm.backupOpenIscsiConfBasedOnShellScript(virtualGuest)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to backup open iscsi conf files from virtual guest `%d`", virtualGuest.Id))
	}

	_, err = vm.writeOpenIscsiInitiatornameBasedOnShellScript(virtualGuest, credential)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi initiatorname from virtual guest `%d`", virtualGuest.Id))
	}

	_, err = vm.writeOpenIscsiConfBasedOnShellScript(virtualGuest, volume, credential)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to write open iscsi conf from virtual guest `%d`", virtualGuest.Id))
	}

	_, err = vm.restartOpenIscsiBasedOnShellScript(virtualGuest)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Failed to restart open iscsi from virtual guest `%d`", virtualGuest.Id))
	}

	_, err = vm.discoveryOpenIscsiTargetsBasedOnShellScript(virtualGuest, volume)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to attach volume with id %d to virtual guest with id: %d.", volume.Id, virtualGuest.Id)
	}

	var deviceName string
	totalTime := time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {
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
	command := fmt.Sprintf("sleep 5; iscsiadm -m discovery -t sendtargets -p %s", volume.ServiceResourceBackendIpAddress)
	_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, command)
	if err != nil {
		return false, bosherr.WrapError(err, "discoverying open iscsi targets")
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

func (vm SoftLayerVM) detachVolumeBasedOnShellScript(virtualGuest datatypes.SoftLayer_Virtual_Guest, volume datatypes.SoftLayer_Network_Storage, hasMultiPath bool) error {
	// umount /var/vcap/store in case read-only mount
	isMounted, err := vm.isMountPoint(virtualGuest, "/var/vcap/store")
	if err != nil {
		return bosherr.WrapError(err, "check mount point /var/vcap/store")
	}

	if isMounted {
		step00 := fmt.Sprintf("umount -l /var/vcap/store")
		_, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step00)
		if err != nil {
			return bosherr.WrapError(err, "umount -l /var/vcap/store")
		}
		vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "umount -l /var/vcap/store", nil)
	}

	// stop open-iscsi
	step1 := fmt.Sprintf("/etc/init.d/open-iscsi stop")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step1)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "/etc/init.d/open-iscsi stop", nil)

	// clean up /etc/iscsi/send_targets/
	step2 := fmt.Sprintf("rm -rf /etc/iscsi/send_targets")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step2)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "rm -rf /etc/iscsi/send_targets", nil)

	// clean up /etc/iscsi/nodes/
	step3 := fmt.Sprintf("rm -rf /etc/iscsi/nodes")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step3)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes")
	}

	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "rm -rf /etc/iscsi/nodes", nil)

	// start open-iscsi
	step4 := fmt.Sprintf("/etc/init.d/open-iscsi start")
	_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step4)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "/etc/init.d/open-iscsi start", nil)

	if hasMultiPath {
		// restart dm-multipath tool
		step5 := fmt.Sprintf("service multipath-tools restart")
		_, err = vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, step5)
		if err != nil {
			return bosherr.WrapError(err, "Restarting Multipath deamon")
		}
		vm.logger.Debug(SOFTLAYER_VM_LOG_TAG, "service multipath-tools restart", nil)
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

func (vm SoftLayerVM) postCheckActiveTransactionsForOSReload(softLayerClient sl.Client) error {
	virtualGuestService, err := softLayerClient.GetSoftLayer_Virtual_Guest_Service()
	if err != nil {
		return bosherr.WrapError(err, "Creating VirtualGuestService from SoftLayer client")
	}

	totalTime := time.Duration(0)
	for totalTime < bslcommon.TIMEOUT {
		activeTransactions, err := virtualGuestService.GetActiveTransactions(vm.ID())
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
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
		return bosherr.WrapError(err, fmt.Sprintf("PowerOn failed with VirtualGuest id %d", vm.ID()))
	}

	vm.logger.Info(SOFTLAYER_VM_OS_RELOAD_TAG, fmt.Sprintf("The virtual guest %d is powered on", vm.ID()))

	return nil
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
			vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "Delete VM transaction started", nil)
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
			vm.logger.Info(SOFTLAYER_VM_LOG_TAG, "VM doesn't exist. Delete done", nil)
			break
		}

		activeTransaction, err := virtualGuestService.GetActiveTransaction(virtualGuestId)
		if err != nil {
			return bosherr.WrapError(err, "Getting active transactions from SoftLayer client")
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
		totalTime += bslcommon.POLLING_INTERVAL
		time.Sleep(bslcommon.POLLING_INTERVAL)
	}

	if totalTime >= bslcommon.TIMEOUT {
		return errors.New(fmt.Sprintf("After deleting a vm, waiting for active transactions to complete TIME OUT!"))
	}

	return nil
}

func (vm SoftLayerVM) isMountPoint(virtualGuest datatypes.SoftLayer_Virtual_Guest, path string) (bool, error) {
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

func (vm SoftLayerVM) searchMounts(virtualGuest datatypes.SoftLayer_Virtual_Guest) ([]Mount, error) {
	var mounts []Mount
	stdout, err := vm.sshClient.ExecCommand(ROOT_USER_NAME, vm.getRootPassword(virtualGuest), virtualGuest.PrimaryBackendIpAddress, "mount")
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
