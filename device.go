package backive

import (
	"log"
	"os/exec"
	"path"
)

var devsByUUID = "/dev/disk/by-uuid"

// Devices is nothing else than a name to Device type mapping
type Devices map[string]*Device

// Device represents a device, with a name easy to remember and the UUID to identify it, optionally an owner.
type Device struct {
	Name      string `mapstructure:",omitempty"`
	UUID      string `mapstructure:"uuid"`
	OwnerUser string `mapstructure:"owner,omitempty"`
	isMounted bool
}

// Mount will mount a device
func (d *Device) Mount() error {
	log.Printf("Mounting device %s, creating directory if it does not exist.\n", d.Name)
	CreateDirectoryIfNotExists(
		path.Join(config.Settings.SystemMountPoint, d.Name),
	)
	//time.Sleep(3000 * time.Millisecond)
	log.Printf("Executing mount command for %s", d.Name)
	cmd := exec.Command(
		"mount",
		path.Join(devsByUUID, d.UUID),
		path.Join(config.Settings.SystemMountPoint, d.Name),
	)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	log.Printf("Command to execute: %s", cmd.String())
	err := mock_cmd_Run(cmd)
	if err != nil {
		log.Printf("Mounting failed with error %v", err)
		return err
	}
	d.isMounted = true
	return nil
}

// Unmount will unmount a device
func (d *Device) Unmount() error {
	if d.isMounted {
		log.Printf("Unmounting %s", d.Name)
		sync := exec.Command("sync")
		syncErr := mock_cmd_Run(sync)
		if syncErr != nil {
			log.Println(syncErr)
			return syncErr
		}
		cmd := exec.Command(
			"umount",
			path.Join(config.Settings.SystemMountPoint, d.Name),
		)
		log.Printf("About to run: %s", cmd.String())
		err := mock_cmd_Run(cmd)
		if err != nil {
			log.Println(err)
			return err
		}
		d.isMounted = false
	}
	return nil
}

// IsMounted returns the mount state of the device
func (d *Device) IsMounted() bool {
	return d.isMounted
}
