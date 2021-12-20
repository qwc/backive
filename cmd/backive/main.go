package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/spf13/viper"
)

var logfile os.File

func createDirectoryIfNotExists(dir string) {
	if _, err := os.Stat(dir); err == nil {
		//ignore
	} else if os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	} else {
		log.Fatal(err)
	}
}

func setupLogging() {
	logname := "/var/log/backive/backive.log"
	logdir, _ := path.Split(logname)
	createDirectoryIfNotExists(logdir)
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

// Save saves the database
func (d *Database) Save() {
	jsonstr, merr := json.Marshal(d.data)
	if merr != nil {
		panic(merr)
	}
	log.Printf("Writing database output to file: %s", jsonstr)
	saveDir, _ := path.Split(d.path)
	createDirectoryIfNotExists(saveDir)
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
	Name       string `mapstructure:",omitempty"`
	UUID       string `mapstructure:"uuid"`
	OwnerUser  string `mapstructure:"owner,omitempty"`
	isMounted  bool
	devsByUuid string "/dev/disk/by-uuid/"
}

// Mount will mount a device
func (d *Device) Mount() error {
	createDirectoryIfNotExists(config.Settings.SystemMountPoint)
	cmd := exec.Command(
		"mount",
		path.Join(d.devsByUuid, d.UUID),
		path.Join(config.Settings.SystemMountPoint, d.Name),
	)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
		return err
	}
	d.isMounted = true
	return nil
}

// Unmount will unmount a device
func (d *Device) Unmount() error {
	sync := exec.Command("sync")
	syncErr := sync.Run()
	if syncErr != nil {
		log.Fatal(syncErr)
		return syncErr
	}
	cmd := exec.Command(
		"umount",
		path.Join(config.Settings.SystemMountPoint, d.Name),
	)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
		return err
	}
	d.isMounted = false
	return nil
}

func (d *Device) IsMounted() bool {
	return d.isMounted
}

// Backup contains all necessary information for executing a configured backup.
type Backup struct {
	Name         string `mapstructure:",omitempty"`
	TargetDevice string `mapstructure:"targetDevice"`
	TargetPath   string `mapstructure:"targetPath"`
	SourcePath   string `mapstructure:"sourcePath"`
	ScriptPath   string `mapstructure:"scriptPath"`
	Frequency    int    `mapstructure:"frequency"`
	ExeUser      string `mapstructure:"user,omitempty"`
	logger       *log.Logger
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
	LogLocation        string `mapstructure:"logLocation"`
	DbLocation         string `mapstructure:"dbLocation"`
}

// Devices is nothing else than a name to Device type mapping
type Devices map[string]Device

// Backups is nothing else than a name to Backup type mapping
type Backups map[string]Backup

func (bs *Backups) findBackupForDevice(d Device) (*Backup, bool) {
	for _, b := range *bs {
		if d.Name == b.TargetDevice {
			return &b, true
		}
	}
	return nil, false
}

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
	log.Println("Initializing EventHandler...")
	var err error
	dir, _ := path.Split(socketPath)
	createDirectoryIfNotExists(dir)
	eh.ls, err = net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	eh.callbacks = make([]func(map[string]string), 3)
}

// Listen starts the event loop.
func (eh *EventHandler) Listen() {
	log.Println("Running eventloop")
	func() {
		for {
			eh.process()
		}
	}()
}

// RegisterCallback adds a function to the list of callback functions for processing of events.
func (eh *EventHandler) RegisterCallback(cb func(map[string]string)) {
	eh.callbacks = append(eh.callbacks, cb)
}

// process processes each and every unix socket event, Unmarshals the json data and calls the list of callbacks.
func (eh *EventHandler) process() {
	client, err := eh.ls.Accept()
	log.Println("Accepted client")
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
	sdata := string(bytes.Trim(data, "\x00"))
	log.Println(sdata)
	env := map[string]string{}
	errjson := json.Unmarshal([]byte(sdata), &env)
	if errjson != nil {
		log.Fatal(errjson)
	}
	for _, v := range eh.callbacks {
		if v != nil {
			v(env)
		}
	}
}

func (b *Backup) CanRun() error {
	// target path MUST exist
	if b.TargetPath == "" {
		return fmt.Errorf("The setting targetPath MUST exist within a backup configuration.")
	}
	//  script must exist, having only script means this is handled in the script
	if b.ScriptPath == "" {
		return fmt.Errorf("The setting scriptPath must exist within a backup configuration.")
	}
	return nil
}

func (b *Backup) PrepareRun() error {
	createDirectoryIfNotExists(path.Join(
		config.Settings.SystemMountPoint,
		b.TargetDevice,
		b.TargetPath,
	))
	// configure extra logger
	logname := "/var/log/backive/backive.log"
	logdir, _ := path.Split(logname)
	createDirectoryIfNotExists(logdir)
	logname = path.Join(logdir, b.Name)
	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalln("Error creating logfile!")
	}
	writer := io.MultiWriter(logfile)
	b.logger = log.New(writer, b.Name, log.LstdFlags)
	return nil
}

// Run runs the backup script with appropriate rights.
func (b *Backup) Run() error {
	if dev, ok := config.Devices[b.Name]; ok && dev.IsMounted() {
		// setup script environment including user to use
		cmd := exec.Command("/usr/bin/sh", b.ScriptPath)
		b.logger.Printf("Running backup script of '%s'", b.Name)
		// does this work?
		cmd.Stdout = b.logger.Writer()
		cmd.Stderr = b.logger.Writer()
		// run script
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Backup '%s' run failed", b.Name)
			return err
		}
		return nil
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

func defaultCallback(envMap map[string]string) {
	if action, ok := envMap["ACTION"]; ok && action == "add" {
		var dev Device
		var uuid string
		if fs_uuid, ok := envMap["ID_FS_UUID"]; !ok {
			log.Fatalln("ID_FS_UUID not available ?!")
		} else {
			uuid = fs_uuid
		}
		log.Println("Device Added")
		var uuidFound bool
		// Check the devices if the UUID is in the config
		for _, device := range config.Devices {
			if uuid == device.UUID {
				uuidFound = true
				dev = device
			}
		}
		if uuidFound {
			dev.Mount()
			backup, found := config.Backups.findBackupForDevice(dev)
			if found {
				err := backup.CanRun()
				if err == nil {
					prepErr := backup.PrepareRun()
					if prepErr != nil {
						log.Fatalf("Error running the backup routine: %v", err)
					}
					rerr := backup.Run()
					if rerr != nil {
						log.Fatalf("Error running the backup routine: %v", err)
					}
				} else {
					log.Fatalf("Error running the backup routine: %v", err)
				}
			}
			dev.Unmount()
		}

	}
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
	events.RegisterCallback(defaultCallback)
	// accept events
	events.Listen()

	// cleanup if anything is there to cleanup
	database.Save()
	log.Println("backive shuting down.")
}
