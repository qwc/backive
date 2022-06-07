package backive

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Configuration struct holding the settings and config items of devices and backups
type Configuration struct {
	Settings Settings `mapstructure:"settings"`
	Devices  Devices  `mapstructure:"devices"`
	Backups  Backups  `mapstructure:"backups"`
	Vconfig  *viper.Viper
}

// Settings struct holds the global configuration items
type Settings struct {
	SystemMountPoint     string `mapstructure:"systemMountPoint"`
	UserMountPoint       string `mapstructure:"userMountPoint"`
	UnixSocketLocation   string `mapstructure:"unixSocketLocation"`
	UIUnixSocketLocation string `mapstructure:"uiUnixSocketLocation"`
	LogLocation          string `mapstructure:"logLocation"`
	DbLocation           string `mapstructure:"dbLocation"`
}

// CreateViper creates a viper instance for usage later
func (c *Configuration) CreateViper() {
	if c.Vconfig == nil {
		vconfig := viper.New()
		//	vconfig.Debug()
		vconfig.SetConfigName("backive")
		// do not set config file explicitly or viper doesnt search for it, and /etc search fails
		//vconfig.SetConfigFile("backive.yml")
		//vconfig.SetConfigFile("backive.yaml")
		vconfig.SetConfigType("yaml")
		//vconfig.AddConfigPath("$HOME/.backive/")
		vconfig.AddConfigPath(".")             // backup config in local dir
		vconfig.AddConfigPath("/etc/backive/") // system config
		vconfig.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Event: %s", e)
			if e.Op == fsnotify.Write {
				log.Printf("Reloading %s", e.Name)
				c.Load()
			}
		})
		vconfig.WatchConfig()
		c.Vconfig = vconfig
	}
}

// Load loads the configuration from the disk
func (c *Configuration) Load() {
	c.CreateViper()
	vc := c.Vconfig
	if err := vc.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("fatal: No config file could be found: %w", err))
		}
		panic(fmt.Errorf("fatal error config file: %w ", err))
	}
	log.Printf("Configuration file used: %s", vc.ConfigFileUsed())

	//Unmarshal all into Configuration type
	err := vc.Unmarshal(c)
	if err != nil {
		fmt.Printf("Error occured when loading config: %v\n", err)
		panic("No configuration available!")
	}
	for k, v := range c.Backups {
		log.Printf("Initializing backup '%s'\n", k)
		v.Name = k
		log.Printf("Initialized backup '%s'\n", v.Name)
	}
	for k, v := range c.Devices {
		log.Printf("Initializing device '%s'\n", k)
		v.Name = k
		log.Printf("Initialized device '%s'\n", v.Name)
	}
}
