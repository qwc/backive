package config

// test package for config

import (
	"bytes"
	"fmt"
	"testing"
)

func TestDummyConfig(t *testing.T) {

	v := CreateViper()
	var yamlExample = []byte(`
settings:
    systemMountPoint: /media/backive
    userMountPoint: $HOME/.backive/mounts
devices:
	my_device:
	    uuid: 98237459872398745987
		owner:
backups:
	my_backup:
		targetDevice: my_device
		targetDir: backive_backup
		sourceDir: /home/user123/stuff
		scriptPath: /path/to/script
		frequency: 7  #weekly
`)
	v.ReadConfig(bytes.NewBuffer(yamlExample))
	var theConfig Configuration
	err := v.Unmarshal(&theConfig)
	if err != nil {
		fmt.Errorf("Unable to decode into struct, %v \n", err)
		panic("Failed!")
	}
	fmt.Printf("systemMountpoint is %v \n", theConfig)

}
