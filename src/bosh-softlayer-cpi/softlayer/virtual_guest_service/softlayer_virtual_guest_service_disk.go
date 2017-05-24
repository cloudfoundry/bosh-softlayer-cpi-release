package instance

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bsl "bosh-softlayer-cpi/softlayer/client"

	"bosh-softlayer-cpi/util"
	datatypes "github.com/softlayer/softlayer-go/datatypes"
)

func (vg SoftlayerVirtualGuestService) getRootPassword(instance datatypes.Virtual_Guest) *string {
	passwords := (*instance.OperatingSystem).Passwords
	for _, password := range passwords {
		if *password.Username == rootUser {
			return password.Password
		}
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) AttachEphemeralDisk(id int, diskSize int) error {
	return vg.softlayerClient.AttachSecondDiskToInstance(id, diskSize)
}

func (vg SoftlayerVirtualGuestService) AttachDisk(id int, diskID int) (string, string, error) {
	volume, err := vg.softlayerClient.GetBlockVolumeDetails(diskID, bsl.VOLUME_DETAIL_MASK)
	if err != nil {
		return "", "", bosherr.WrapErrorf(err, "Fetching volume details with id '%d'", diskID)
	}

	instance, err := vg.softlayerClient.GetInstance(id, bsl.INSTANCE_DETAIL_MASK)
	if err != nil {
		return "", "", bosherr.WrapErrorf(err, "Fetching instance details with id '%d'", id)
	}

	password := vg.getRootPassword(instance)
	if password == nil {
		return "", "", bosherr.WrapErrorf(err, "Failed to retrieve root password with id '%d'", id)
	}

	until := time.Now().Add(time.Duration(1) * time.Hour)
	err = vg.softlayerClient.AuthorizeHostToVolume(&instance, diskID, until)
	if err != nil {
		return "", "", bosherr.WrapErrorf(err, "Authorizing vm with id '%d' to disk with id '%d'", id, diskID)
	}

	ssh := util.GetSshClient(rootUser, *password, *instance.PrimaryBackendIpAddress)

	deviceName, err := vg.waitForVolumeAttached(id, volume, ssh)
	if err != nil {
		return "", "", bosherr.WrapError(err, fmt.Sprintf("Failed to attach volume '%d' to virtual guest '%d'", diskID, id))
	}

	vg.logger.Info(softlayerVirtualGuestServiceLogTag, "The volume device name '%s', device path '%s'", deviceName, fmt.Sprintf("%s/%s", volumePathPrefix, deviceName))
	return deviceName, fmt.Sprintf("%s/%s", volumePathPrefix, deviceName), nil
}

func (vg SoftlayerVirtualGuestService) AttachedDisks(id int) ([]string, error) {
	return vg.softlayerClient.GetAllowedNetworkStorage(id)
}

func (vg SoftlayerVirtualGuestService) ReAttachLeftDisk(id int, devicePath string, diskID int) error {
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Left Disk Id %d", diskID)
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "Left Disk device path %s", devicePath)

	volume, err := vg.softlayerClient.GetBlockVolumeDetails(diskID, bsl.VOLUME_DETAIL_MASK)
	if err != nil {
		return bosherr.WrapErrorf(err, "Fetching volume details with id '%d'", diskID)
	}

	instance, err := vg.softlayerClient.GetInstance(id, bsl.INSTANCE_DETAIL_MASK)
	if err != nil {
		return bosherr.WrapErrorf(err, "Fetching instance details with id '%d'", id)
	}

	password := vg.getRootPassword(instance)
	if password == nil {
		return bosherr.WrapErrorf(err, "Retrieving root password with id '%d'", id)
	}

	ssh := util.GetSshClient(rootUser, *password, *instance.PrimaryBackendIpAddress)
	_, err = vg.discoveryOpenIscsiTargetsBasedOnShellScript(volume, ssh)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Reattaching volume with id '%d' to instance with id '%d'", diskID, id))
	}

	command := fmt.Sprintf("sleep 5; mount %s-part1 /var/vcap/store", devicePath)
	_, err = ssh.ExecCommand(command)
	if err != nil {
		return bosherr.WrapError(err, "Mounting /var/vcap/store")
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) DetachDisk(id int, diskID int) error {
	volume, err := vg.softlayerClient.GetBlockVolumeDetails(diskID, bsl.VOLUME_DETAIL_MASK)
	if err != nil {
		return bosherr.WrapErrorf(err, "Fetching volume details with id '%d'", diskID)
	}

	instance, err := vg.softlayerClient.GetInstance(id, bsl.INSTANCE_DETAIL_MASK)
	if err != nil {
		return bosherr.WrapErrorf(err, "Fetching instance details with id '%d'", id)
	}

	password := vg.getRootPassword(instance)
	if password == nil {
		return bosherr.WrapErrorf(err, "Retrieving root password with id '%d'", id)
	}

	ssh := util.GetSshClient(rootUser, *password, *instance.PrimaryBackendIpAddress)

	err = vg.detachVolumeBasedOnShellScript(volume, ssh)
	if err != nil {
		return bosherr.WrapErrorf(err, "Detaching volume with id '%d' from virtual guest with id '%d'", diskID, id)
	}

	until := time.Now().Add(time.Duration(1) * time.Hour)
	err = vg.softlayerClient.DeauthorizeHostToVolume(&instance, diskID, until)
	if err != nil {
		return bosherr.WrapErrorf(err, "De-Authorizing vm with id '%d' to disk with id '%d'", id, diskID)
	}

	return nil
}

func (vg SoftlayerVirtualGuestService) waitForVolumeAttached(id int, volume datatypes.Network_Storage, ssh util.SshClient) (string, error) {
	oldDisks, err := vg.getIscsiDeviceNamesBasedOnShellScript(ssh)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Getting devices names from virtual guest '%d'", id))
	}
	if len(oldDisks) > 2 {
		return "", bosherr.Error(fmt.Sprintf("Too manay persistent disks attached to virtual guest '%d'", id))
	}

	credential, err := vg.softlayerClient.GetAllowedHostCredential(id)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Getting iscsi host auth from virtual guest '%d'", id))
	}

	_, err = vg.backupOpenIscsiConfBasedOnShellScript(ssh)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Backuping open iscsi conf files from virtual guest '%d'", id))
	}

	_, err = vg.writeOpenIscsiInitiatornameBasedOnShellScript(credential, ssh)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Writing open iscsi initiatorname from virtual guest '%d'", id))
	}

	_, err = vg.writeOpenIscsiConfBasedOnShellScript(volume, credential, ssh)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Writing open iscsi conf from virtual guest '%d'", id))
	}

	_, err = vg.restartOpenIscsiBasedOnShellScript(ssh)
	if err != nil {
		return "", bosherr.WrapError(err, fmt.Sprintf("Restarting open iscsi from virtual guest '%d'", id))
	}

	_, err = vg.discoveryOpenIscsiTargetsBasedOnShellScript(volume, ssh)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Attaching volume with id %d to virtual guest with id: %d.", *volume.Id, id)
	}

	var deviceName string
	totalTime := time.Duration(0)
	for totalTime < 5*time.Minute {
		newDisks, err := vg.getIscsiDeviceNamesBasedOnShellScript(ssh)
		if err != nil {
			return "", bosherr.WrapError(err, fmt.Sprintf("Getting devices names from virtual guest '%d'", id))
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

	return "", bosherr.Errorf("Attaching disk '%d' to virtual guest '%d'", *volume.Id, id)
}

func (vg SoftlayerVirtualGuestService) getIscsiDeviceNamesBasedOnShellScript(ssh util.SshClient) ([]string, error) {
	devices := []string{}

	command := fmt.Sprintf("dmsetup ls")

	result, err := ssh.ExecCommand(command)
	if err != nil {
		return devices, err
	}
	if strings.Contains(result, "No devices found") {
		return devices, nil
	}
	vg.logger.Info(softlayerVirtualGuestServiceLogTag, fmt.Sprintf("Devices on VM %s", result))
	lines := strings.Split(strings.Trim(result, "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		if match, _ := regexp.MatchString("-part1", lines[i]); !match {
			devices = append(devices, strings.Fields(lines[i])[0])
		}
	}

	return devices, nil
}

func (vg SoftlayerVirtualGuestService) backupOpenIscsiConfBasedOnShellScript(ssh util.SshClient) (bool, error) {
	command := fmt.Sprintf("cp /etc/iscsi/iscsid.conf{,.save}")
	_, err := ssh.ExecCommand(command)
	if err != nil {
		return false, bosherr.WrapError(err, "backuping open iscsi conf")
	}

	return true, nil
}

func (vg SoftlayerVirtualGuestService) restartOpenIscsiBasedOnShellScript(ssh util.SshClient) (bool, error) {
	command := fmt.Sprintf("/etc/init.d/open-iscsi restart")
	_, err := ssh.ExecCommand(command)
	if err != nil {
		return false, bosherr.WrapError(err, "restarting open iscsi")
	}

	return true, nil
}

func (vg SoftlayerVirtualGuestService) discoveryOpenIscsiTargetsBasedOnShellScript(volume datatypes.Network_Storage, ssh util.SshClient) (bool, error) {
	command := fmt.Sprintf("sleep 5; iscsiadm -m discovery -t sendtargets -p %s", *volume.ServiceResourceBackendIpAddress)
	_, err := ssh.ExecCommand(command)
	if err != nil {
		return false, bosherr.WrapError(err, "Discoverying open iscsi targets")
	}

	command = "sleep 5; echo `iscsiadm -m node -l`"
	_, err = ssh.ExecCommand(command)
	if err != nil {
		return false, bosherr.WrapError(err, "login iscsi targets")
	}

	return true, nil
}

func (vg SoftlayerVirtualGuestService) writeOpenIscsiInitiatornameBasedOnShellScript(credential datatypes.Network_Storage_Allowed_Host, ssh util.SshClient) (bool, error) {
	if len(*credential.Name) > 0 {
		command := fmt.Sprintf("echo 'InitiatorName=%s' > /etc/iscsi/initiatorname.iscsi", *credential.Name)
		_, err := ssh.ExecCommand(command)
		if err != nil {
			return false, bosherr.WrapError(err, "Writing to /etc/iscsi/initiatorname.iscsi")
		}
	}

	return true, nil
}

func (vg SoftlayerVirtualGuestService) writeOpenIscsiConfBasedOnShellScript(volume datatypes.Network_Storage, allowedHost datatypes.Network_Storage_Allowed_Host, ssh util.SshClient) (bool, error) {
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

	if err = ssh.UploadFile(file.Name(), "/etc/iscsi/iscsid.conf"); err != nil {
		return false, bosherr.WrapError(err, "Writing to /etc/iscsi/iscsid.conf")
	}

	return true, nil
}

func (vg SoftlayerVirtualGuestService) detachVolumeBasedOnShellScript(volume datatypes.Network_Storage, ssh util.SshClient) error {
	// umount /var/vcap/store in case read-only mount
	isMounted, err := vg.isMountPoint("/var/vcap/store", ssh)
	if err != nil {
		return bosherr.WrapError(err, "check mount point /var/vcap/store")
	}

	if isMounted {
		step00 := fmt.Sprintf("umount -l /var/vcap/store")
		_, err := ssh.ExecCommand(step00)
		if err != nil {
			return bosherr.WrapError(err, "umount -l /var/vcap/store")
		}
		vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "umount -l /var/vcap/store", nil)
	}

	// stop open-iscsi
	step1 := fmt.Sprintf("/etc/init.d/open-iscsi stop")
	_, err = ssh.ExecCommand(step1)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "/etc/init.d/open-iscsi stop", nil)

	// clean up /etc/iscsi/send_targets/
	step2 := fmt.Sprintf("rm -rf /etc/iscsi/send_targets")
	_, err = ssh.ExecCommand(step2)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/send_targets")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "rm -rf /etc/iscsi/send_targets", nil)

	// clean up /etc/iscsi/nodes/
	step3 := fmt.Sprintf("rm -rf /etc/iscsi/nodes")
	_, err = ssh.ExecCommand(step3)
	if err != nil {
		return bosherr.WrapError(err, "Removing /etc/iscsi/nodes")
	}

	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "rm -rf /etc/iscsi/nodes", nil)

	// start open-iscsi
	step4 := fmt.Sprintf("/etc/init.d/open-iscsi start")
	_, err = ssh.ExecCommand(step4)
	if err != nil {
		return bosherr.WrapError(err, "Restarting open iscsi")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "/etc/init.d/open-iscsi start", nil)

	// restart dm-multipath tool
	step5 := fmt.Sprintf("service multipath-tools restart")
	_, err = ssh.ExecCommand(step5)
	if err != nil {
		return bosherr.WrapError(err, "Restarting Multipath deamon")
	}
	vg.logger.Debug(softlayerVirtualGuestServiceLogTag, "service multipath-tools restart", nil)

	return nil
}

func (vg SoftlayerVirtualGuestService) isMountPoint(path string, ssh util.SshClient) (bool, error) {
	mounts, err := vg.searchMounts(ssh)
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

func (vg SoftlayerVirtualGuestService) searchMounts(ssh util.SshClient) ([]Mount, error) {
	var mounts []Mount
	stdout, err := ssh.ExecCommand("mount")
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
