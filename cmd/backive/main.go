package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
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
var devsByUuid string = "/dev/disk/by-uuid"
var dbPath string = "/var/lib/backive/data.json"

// Database is a simple string to string mapping, where arbitrary strings can be stored and safed to disk or loaded
type Database struct {
	data map[string]string
}

// Save saves the database
func (d *Database) Save() {
	jsonstr, merr := json.Marshal(d.data)
	if merr != nil {
		panic(merr)
	}
	log.Printf("Writing database output to file: %s", jsonstr)
	saveDir, _ := path.Split(dbPath)
	createDirectoryIfNotExists(saveDir)
	err := os.WriteFile(dbPath, []byte(jsonstr), 0644)
	if err != nil {
		panic(err)
	}
}

// LoadDb loads the database
func (d *Database) Load() {
	if _, err := os.Stat(dbPath); err == nil {
		data, rferr := os.ReadFile(dbPath)
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
func (d *Device) Mount() error {
	log.Printf("Mounting device %s, creating directory if it does not exist.\n", d.Name)
	createDirectoryIfNotExists(
		path.Join(config.Settings.SystemMountPoint, d.Name),
	)
	//time.Sleep(3000 * time.Millisecond)
	log.Printf("Executing mount command for %s", d.Name)
	cmd := exec.Command(
		"mount",
		path.Join(devsByUuid, d.UUID),
		path.Join(config.Settings.SystemMountPoint, d.Name),
	)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	log.Printf("Command to execute: %s", cmd.String())
	err := cmd.Run()
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
		syncErr := sync.Run()
		if syncErr != nil {
			log.Println(syncErr)
			return syncErr
		}
		cmd := exec.Command(
			"umount",
			path.Join(config.Settings.SystemMountPoint, d.Name),
		)
		log.Printf("About to run: %s", cmd.String())
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			return err
		}
		d.isMounted = false
	}
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
	Vconfig  *viper.Viper
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
type Devices map[string]*Device

// Backups is nothing else than a name to Backup type mapping
type Backups map[string]*Backup

// findBackupsForDevice only finds the first backup which is configured for a given device.
func (bs *Backups) findBackupsForDevice(d Device) ([]*Backup, bool) {
	var backups []*Backup = []*Backup{}
	for _, b := range *bs {
		if d.Name == b.TargetDevice {
			backups = append(backups, b)
		}
	}
	var ret bool = len(backups) > 0
	return backups, ret
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
	c.Vconfig = vconfig
}

// Load loads the configuration from the disk
func (c *Configuration) Load() {
	c.CreateViper()
	vc := c.Vconfig
	if err := vc.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("Fatal: No config file could be found: %w", err))
		}
		panic(fmt.Errorf("Fatal error config file: %w ", err))
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
	//log.Println(sdata)
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
	if !b.ShouldRun() {
		return fmt.Errorf("Frequency (days inbetween) not reached.")
	}
	return nil
}

func (b *Backup) PrepareRun() error {
	backupPath := path.Join(
		config.Settings.SystemMountPoint,
		b.TargetDevice,
		b.TargetPath,
	)
	createDirectoryIfNotExists(backupPath)
	// configure extra logger
	logname := "/var/log/backive/backive.log"
	logdir, _ := path.Split(logname)
	createDirectoryIfNotExists(logdir)
	logname = path.Join(logdir, b.Name) + ".log"
	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Println("Error creating logfile!")
		return err
	}
	writer := io.MultiWriter(logfile)
	b.logger = log.New(writer, b.Name, log.LstdFlags)
	cmd := exec.Command("chown", "-R", b.ExeUser, backupPath)
	err = cmd.Run()
	if err != nil {
		b.logger.Printf("chown for backup directory failed: %s", err)
		return err
	}
	return nil
}

// Run runs the backup script with appropriate rights.
func (b *Backup) Run() error {
	log.Printf("Running backup '%s'.", b.Name)
	dev, ok := config.Devices[b.TargetDevice]
	if ok {
		log.Printf("Device found: %s (%s).", dev.Name, dev.UUID)
	} else {
		log.Printf("Device %s not found", b.TargetDevice)
	}
	if ok && dev.IsMounted() {
		if !strings.ContainsAny(b.ScriptPath, "/") || strings.HasPrefix(b.ScriptPath, ".") {
			//The scriptPath is a relative path, from the place of the config, so use the config as base
			log.Printf("ERROR: Script path is relative, aborting.")
			return fmt.Errorf("Script path is relative, aborting.")
		}
		cmd := exec.Command("/usr/bin/sh", b.ScriptPath)
		if b.ExeUser != "" {
			// setup script environment including user to use
			cmd = exec.Command("sudo", "-E", "-u", b.ExeUser, "/usr/bin/sh", b.ScriptPath)
		}
		b.logger.Printf("Running backup script of '%s'", b.Name)
		b.logger.Printf("Script is: %s", b.ScriptPath)
		b.logger.Printf("Full command is: %s", cmd.String())
		cmd.Stdout = b.logger.Writer()
		cmd.Stderr = b.logger.Writer()
		cmd.Env = []string{
			fmt.Sprintf("BACKIVE_MOUNT=%s", config.Settings.SystemMountPoint),
			fmt.Sprintf("BACKIVE_TO=%s",
				path.Join(config.Settings.SystemMountPoint, dev.Name, b.TargetPath),
			),
			fmt.Sprintf("BACKIVE_FROM=%s", b.SourcePath),
		}
		log.Printf("Environment for process: %s", cmd.Env)
		cmd.Dir = path.Join(config.Settings.SystemMountPoint, dev.Name)

		log.Printf("About to run: %s", cmd.String())
		// run script
		err := cmd.Run()
		if err != nil {
			log.Printf("Backup '%s' run failed", b.Name)
			return err
		}
		runs.RegisterRun(b)
		return nil
	}
	// quit with error that the device is not available.
	return fmt.Errorf("The device is not mounted")
}

// Runs contains the Data for the scheduler: mapping from backups to a list of timestamps of the last 10 backups
type Runs struct {
	data map[string][]time.Time
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
		if days >= float64(freq) {
			return true
		}
	}
	if freq == 0 {
		return true
	}
	return false
}

// RegisterRun saves a date of a backup run into the internal storage
func (r *Runs) RegisterRun(b *Backup) {
	if r.data == nil {
		r.data = map[string][]time.Time{}
	}
	nbl, ok := r.data[b.Name]
	if !ok {
		nbl = make([]time.Time, 1)
	}
	nbl = append([]time.Time{time.Now()}, nbl...)
	r.data[b.Name] = nbl
	r.Save(database)
}

// LastRun returns the time.Time of the last run of the backup given.
func (r *Runs) LastRun(b Backup) (time.Time, error) {
	_, ok := r.data[b.Name]
	if ok {
		slice := r.data[b.Name]
		if len(slice) > 0 {
			var t = time.Time(slice[0])
			return t, nil
		}
	}
	return time.Unix(0, 0), fmt.Errorf("Backup name not found and therefore has never run")
}

func defaultCallback(envMap map[string]string) {
	if action, ok := envMap["ACTION"]; ok && action == "add" {
		var dev *Device
		var uuid string
		if fs_uuid, ok := envMap["ID_FS_UUID"]; !ok {
			log.Println("ID_FS_UUID not available ?!")
			return
		} else {
			uuid = fs_uuid
		}
		log.Println("Device connected.")
		var uuidFound bool
		// Check the devices if the UUID is in the config
		for _, device := range config.Devices {
			if uuid == device.UUID {
				uuidFound = true
				dev = device
			}
		}
		if uuidFound {
			log.Println("Device recognized.")
			log.Printf("Device: Name: %s, UUID: %s", dev.Name, dev.UUID)
			backups, found := config.Backups.findBackupsForDevice(*dev)
			log.Println("Searching configured backups...")
			if found {
				for _, backup := range backups {
					log.Printf("Backup found: %s", backup.Name)
					err := backup.CanRun()
					if err == nil {
						// only mount device if we really have to do a backup!
						dev.Mount()
						log.Println("Device mounted.")
						log.Println("Backup is able to run (config check passed).")
						prepErr := backup.PrepareRun()
						log.Println("Prepared run.")
						if prepErr != nil {
							log.Printf("Error running the backup routine: %v", err)
						}
						log.Println("Running backup.")
						rerr := backup.Run()
						if rerr != nil {
							log.Printf("Error running the backup routine: %v", err)
						}
						dev.Unmount()
					} else {
						log.Printf("Backup '%s' can not run (error or frequency not reached): %s", backup.Name, err)
					}
				}
			} else {
				log.Println("No backup found.")
			}
		}

	}
}

func main() {
	setupLogging()
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	exit_chan := make(chan int)
	go func() {
		for {
			s := <-signal_chan
			switch s {
			case syscall.SIGHUP:
				log.Println("hungup")
			case syscall.SIGINT:
				log.Println("Ctrl+C, quitting.")
				exit_chan <- 0
			case syscall.SIGTERM:
				log.Println("Terminating.")
				exit_chan <- 0
			case syscall.SIGQUIT:
				log.Println("Quitting")
				exit_chan <- 0
			default:
				log.Println("Unknown signal.")
				exit_chan <- 1
			}
		}
	}()
	go func() {
		// exit function only does something when the exit_chan has an item
		// cleaning up stuff
		code := <-exit_chan
		database.Save()
		log.Printf("Received exit code (%d), shutting down.", code)
		os.Exit(code)
	}()

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
}
