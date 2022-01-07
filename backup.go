package backive

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// Backup contains all necessary information for executing a configured backup.
type Backup struct {
	Name         string `mapstructure:",omitempty"`
	TargetDevice string `mapstructure:"targetDevice"`
	TargetPath   string `mapstructure:"targetPath"`
	SourcePath   string `mapstructure:"sourcePath"`
	ScriptPath   string `mapstructure:"scriptPath"`
	Frequency    int    `mapstructure:"frequency"`
	ExeUser      string `mapstructure:"user,omitempty"`
	Label        string `mapstructure:"label,omitempty"`
	logger       *log.Logger
}

// Backups is nothing else than a name to Backup type mapping
type Backups map[string]*Backup

// FindBackupsForDevice only finds the first backup which is configured for a given device.
func (bs *Backups) FindBackupsForDevice(d Device) ([]*Backup, bool) {
	var backups = []*Backup{}
	for _, b := range *bs {
		if d.Name == b.TargetDevice {
			backups = append(backups, b)
		}
	}
	var ret = len(backups) > 0
	return backups, ret
}

// CanRun Checks the configuration items required and checks the frequency setting with the run database if a Backup should run.
func (b *Backup) CanRun() error {
	// target path MUST exist
	if b.TargetPath == "" {
		return fmt.Errorf("the setting targetPath MUST exist within a backup configuration")
	}
	//  script must exist, having only script means this is handled in the script
	if b.ScriptPath == "" {
		return fmt.Errorf("the setting scriptPath must exist within a backup configuration")
	}
	if !b.ShouldRun() {
		return fmt.Errorf("frequency (days inbetween) not reached")
	}
	return nil
}

// PrepareRun prepares a run for a backup, creates a logger for the execution of the backup script and gives the rights of the directory recursively to the user specified.
func (b *Backup) PrepareRun() error {
	backupPath := path.Join(
		config.Settings.SystemMountPoint,
		b.TargetDevice,
		b.TargetPath,
	)
	CreateDirectoryIfNotExists(backupPath)
	// configure extra logger
	logname := "/var/log/backive/backive.log"
	logdir, _ := path.Split(logname)
	CreateDirectoryIfNotExists(logdir)
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
			return fmt.Errorf("script path is relative, aborting")
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
	return fmt.Errorf("the device is not mounted")
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
	lr, ok := runs.LastRun(b)
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
func (r *Runs) LastRun(b *Backup) (time.Time, error) {
	_, ok := r.data[b.Name]
	if ok {
		slice := r.data[b.Name]
		if len(slice) > 0 {
			var t = time.Time(slice[0])
			return t, nil
		}
	}
	return time.Unix(0, 0), fmt.Errorf("backup name not found and therefore has never run")
}
