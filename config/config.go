package config

import (
	"fmt"

	"github.com/qwc/backive/core"

	"github.com/spf13/viper"
)

type Configuration struct {
	Settings Settings `mapstructure:"settings"`
	Devices  Devices  `mapstructure:"devices"`
	Backups  Backups  `mapstructure:"backups"`
}

type Settings struct {
	SystemMountPoint string `mapstructure:"systemMountPoint"`
	UserMountPoint   string `mapstructure:"userMountPoint"`
}

type Devices map[string]core.Device

type Backups map[string]core.Backup

func CreateViper() *viper.Viper {
	vconfig := viper.New()
	vconfig.SetConfigName("backive")
	vconfig.SetConfigType("yaml")
	vconfig.AddConfigPath("/etc/backive/") // system config
	vconfig.AddConfigPath("$HOME/.backive/")
	vconfig.AddConfigPath(".")
	return vconfig
}

func Load() *Configuration {
	vconfig := CreateViper()
	if err := vconfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("Fatal: No config file could be found!"))
		}
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	var cfg *Configuration

	err := vconfig.Unmarshal(cfg)
	if err != nil {
		fmt.Printf("Error occured when loading config: %v\n", err)
		panic("No configuration available!")
	}
	//Unmarshal all into Configuration type
	return cfg
}
