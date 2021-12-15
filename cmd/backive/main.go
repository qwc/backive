package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/viper"
)

var logfile os.File

func setupLogging() {
	logdir := "/var/log/backive"
	logname := "/var/log/backive/backive.log"
	if _, err := os.Stat(logdir); err == nil {
		//ignore
	} else if os.IsNotExist(err) {
		os.MkdirAll(logdir, 0755)
	}
	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error creating logfile!")
		panic("no logfile no info")
	}
	log.SetOutput(logfile)
}

// Global variables for backive
var (
	database Database
	config   Configuration
	runs     Runs
	events   EventHandler
)

// Database is a simple string to string mapping, where arbitrary strings can be stored and safed to disk or loaded
type Database struct {
	data map[string]string
	path string "/var/lib/backive/data.json"
}

// SaveDb saves the database
func (d *Database) Save() {
	jsonstr, merr := json.Marshal(d.data)
	if merr != nil {
		panic(merr)
	}

	err := os.WriteFile(d.path, []byte(jsonstr), 0644)
	if err != nil {
		panic(err)
	}
}

// LoadDb loads the database
func (d *Database) Load() {
	if _, err := os.Stat(d.path); err == nil {
		data, rferr := os.ReadFile(d.path)
		if rferr != nil {
			panic(rferr)
		}
		json.Unmarshal(data, &d.data)
	} else if os.IsNotExist(err) {
		// no data

	}
}

// Device represents a device, with a name easy to remember and the UUID to identify it, optionally an owner.
type Device struct {
	Name      string `mapstructure:",omitempty"`
	UUID      string `mapstructure:"uuid"`
	OwnerUser string `mapstructure:"owner,omitempty"`
	isMounted bool
}

// Mount will mount a device
func (d *Device) Mount() {

	d.isMounted = true
}

// Unmount will unmount a device
func (d *Device) Unmount() {

	d.isMounted = false
}

func (d *Device) IsMounted() bool {
	return d.isMounted
}

// Backup contains all necessary information for executing a configured backup.
type Backup struct {
	Name         string `mapstructure:",omitempty"`
	TargetDevice string `mapstructure:"targetDevice"`
	TargetDir    string `mapstructure:"targetDir"`
	SourceDir    string `mapstructure:"sourceDir"`
	ScriptPath   string `mapstructure:"scriptPath"`
	Frequency    int    `mapstructure:"frequency"`
	ExeUser      string `mapstructure:"user,omitempty"`
}

// Configuration struct holding the settings and config items of devices and backups
type Configuration struct {
	Settings Settings `mapstructure:"settings"`
	Devices  Devices  `mapstructure:"devices"`
	Backups  Backups  `mapstructure:"backups"`
	vconfig  *viper.Viper
}

// Settings struct holds the global configuration items
type Settings struct {
	SystemMountPoint   string `mapstructure:"systemMountPoint"`
	UserMountPoint     string `mapstructure:"userMountPoint"`
	UnixSocketLocation string `mapstructure:"unixSocketLocation"`
}

// Devices is nothing else than a name to Device type mapping
type Devices map[string]Device

// Backups is nothing else than a name to Backup type mapping
type Backups map[string]Backup

// CreateViper creates a viper instance for usage later
func (c *Configuration) CreateViper() {
	vconfig := viper.New()
	//	vconfig.Debug()
	vconfig.SetConfigName("backive")
	vconfig.SetConfigFile("backive.yml")
	//vconfig.SetConfigFile("backive.yaml")
	vconfig.SetConfigType("yaml")
	vconfig.AddConfigPath("/etc/backive/") // system config
	//vconfig.AddConfigPath("$HOME/.backive/")
	vconfig.AddConfigPath(".")
	c.vconfig = vconfig
}

// Load loads the configuration from the disk
func (c *Configuration) Load() {
	c.CreateViper()
	vc := c.vconfig
	if err := vc.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("Fatal: No config file could be found: %w", err))
		}
		panic(fmt.Errorf("Fatal error config file: %w ", err))
	}

	//Unmarshal all into Configuration type
	err := vc.Unmarshal(c)
	if err != nil {
		fmt.Printf("Error occured when loading config: %v\n", err)
		panic("No configuration available!")
	}
	for k, v := range c.Backups {
		v.Name = k
	}
	for k, v := range c.Devices {
		v.Name = k
	}
}

type EventHandler struct {
	ls        net.Listener
	done      <-chan struct{}
	callbacks []func(map[string]string)
}

// Init initializes the unix socket.
func (eh *EventHandler) Init(socketPath string) {
	var err error
	eh.ls, err = net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	eh.callbacks = make([]func(map[string]string), 3)
}

// Listen starts the event loop.
func (eh *EventHandler) Listen() {
	for {
		go func() {
			eh.process()
		}()
	}
}

// RegisterCallback adds a function to the list of callback functions for processing of events.
func (eh *EventHandler) RegisterCallback(cb func(map[string]string)) {
	eh.callbacks = append(eh.callbacks, cb)
}

// process processes each and every unix socket event, Unmarshals the json data and calls the list of callbacks.
func (eh *EventHandler) process() {
	client, err := eh.ls.Accept()
	if err != nil {
		log.Fatal(err)
	}
	data := make([]byte, 2048)
	for {
		buf := make([]byte, 512)
		nr, err := client.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		data = append(data, buf[0:nr]...)
		if err == io.EOF {
			break
		}
	}
	log.Println(data)
	env := map[string]string{}
	errjson := json.Unmarshal(data, &env)
	if errjson != nil {
		log.Fatal(errjson)
	}
	for _, v := range eh.callbacks {
		v(env)
	}
}

// Run runs the backup script with appropriate rights.
func (b *Backup) Run() error {
	cfg := config
	if dev := cfg.Devices[b.Name]; dev.IsMounted() {
		checkExistence := func(path string, name string) error {
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("%s does not exist", name)
				} else {
					return fmt.Errorf("Error when checking %s: %w", name, err)
				}
			}
			return nil
		}
		// Check for existence of target dir
		if err := checkExistence(b.TargetDir, "target directory"); err != nil {
			return err
		}
		// check for existence of source dir
		if err := checkExistence(b.SourceDir, "source directory"); err != nil {
			return err
		}
		// check for existence of script path
		if err := checkExistence(b.ScriptPath, "script path"); err != nil {
			return err
		}
		// setup script environment including user to use
		// run script
	}
	// quit with error that the device is not available.
	return fmt.Errorf("The device is not mounted")
}

type backupRuns struct {
	runlist *list.List
}

// Runs contains the Data for the scheduler: mapping from backups to a list of timestamps of the last 10 backups
type Runs struct {
	data map[string]backupRuns
}

// Load loads the data from the json database
func (r *Runs) Load(db Database) {
	data := db.data["runs"]
	if data != "" {
		runerr := json.Unmarshal([]byte(db.data["runs"]), &r.data)
		if runerr != nil {
			panic(runerr)
		}
	}
}

// Save saves the data into the json database
func (r *Runs) Save(db Database) {
	str, err := json.Marshal(r.data)
	if err != nil {
		panic(err)
	}
	db.data["runs"] = string(str)
}

// ShouldRun Takes a backup key and returns a bool if a backup should run now.
func (b *Backup) ShouldRun() bool {
	freq := b.Frequency
	// calculate time difference from last run, return true if no run has taken place
	lr, ok := runs.LastRun(*b)
	if ok == nil {
		dur := time.Since(lr)
		days := dur.Hours() / 24
		if days > float64(freq) {
			return true
		}
	}
	if freq == 0 {
		return true
	}
	return false
}

// RegisterRun saves a date of a backup run into the internal storage
func (r *Runs) RegisterRun(b Backup) {
	nbl, ok := r.data[b.Name]
	if !ok {
		nbl.runlist = list.New()
		r.data[b.Name] = nbl
	}
	nbl.runlist.PushFront(time.Now())
	r.Save(database)
}

// LastRun returns the time.Time of the last run of the backup given.
func (r *Runs) LastRun(b Backup) (time.Time, error) {
	_, ok := r.data[b.Name]
	if ok {
		var t = time.Time(r.data[b.Name].runlist.Front().Value.(time.Time))
		return t, nil
	}
	return time.Unix(0, 0), fmt.Errorf("Backup name not found and therefore has never run")
}

func main() {
	setupLogging()
	// TODO: do proper signal handling!
	log.Println("backive starting up...")
	// find and load config
	database.Load()
	config.Load()
	runs.Load(database)

	// init scheduler and check for next needed runs?

	// start event loop
	events.Init(config.Settings.UnixSocketLocation)
	// accept event
	events.Listen()

	// cleanup if anything is there to cleanup
	database.Save()
	log.Println("backive shuting down.")
}
