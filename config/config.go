package config

import (
	"fmt"

	"github.com/qwc/backive/backup"
	"github.com/qwc/backive/device"

	"github.com/spf13/viper"
)

var (
	config  *Configuration
	vconfig *viper.Viper
)

// Configuration struct holding the settings and config items of devices and backups
type Configuration struct {
	Settings Settings `mapstructure:"settings"`
	Devices  Devices  `mapstructure:"devices"`
	Backups  Backups  `mapstructure:"backups"`
}

// Settings struct holds the global configuration items
type Settings struct {
	SystemMountPoint string `mapstructure:"systemMountPoint"`
	UserMountPoint   string `mapstructure:"userMountPoint"`
}

// Devices is nothing else than a name to Device type mapping
type Devices map[string]device.Device

// Backups is nothing else than a name to Backup type mapping
type Backups map[string]backup.Backup

// CreateViper creates a viper instance for usage later
func CreateViper() *viper.Viper {
	vconfig := viper.New()
	vconfig.SetConfigName("backive")
	vconfig.SetConfigType("yaml")
	vconfig.AddConfigPath("/etc/backive/") // system config
	vconfig.AddConfigPath("$HOME/.backive/")
	vconfig.AddConfigPath(".")
	return vconfig
}

// Load loads the configuration from the disk
func Load() *Configuration {
	vconfig := CreateViper()
	if err := vconfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("Fatal: No config file could be found"))
		}
		panic(fmt.Errorf("Fatal error config file: %w ", err))
	}

	var cfg *Configuration

	//Unmarshal all into Configuration type
	err := vconfig.Unmarshal(cfg)
	if err != nil {
		fmt.Printf("Error occured when loading config: %v\n", err)
		panic("No configuration available!")
	}
	for k, v := range cfg.Backups {
		v.Name = k
	}
	for k, v := range cfg.Devices {
		v.Name = k
	}
	return cfg
}

// Init Initializes the configuration
func Init() {
	config = Load()
}

// Get returns the Configuration global variable
func Get() *Configuration {
	return config
}
