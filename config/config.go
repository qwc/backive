package config

import "github.com/spf13/viper"

type Device struct {
	name      string
	uuid      string
	ownerUser string
}

type Backup struct {
	name             string
	targetDeviceName string
	targetDir        string
	sourceDir        string
	scriptPath       string
	frequency        int
	exeUser          string
}

type Settings struct {
	systemMountPoint string
	userMountPoint   string
}

func loadDevice() {
	v1 := viper.New()
	v1.SetConfigName("devices")

}
