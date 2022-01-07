package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/qwc/backive"
)

func setupLogging() {
	logname := path.Join(config.Settings.LogLocation, "backive.log")
	logdir, _ := path.Split(logname)
	backive.CreateDirectoryIfNotExists(logdir)
	logfile, err := os.OpenFile(logname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error creating logfile!")
		panic("no logfile no info")
	}
	log.SetOutput(logfile)
}

// Global variables for backive
var (
	config   backive.Configuration
	database backive.Database
	events   backive.EventHandler
)

func defaultCallback(envMap map[string]string) {
	if action, ok := envMap["ACTION"]; ok && action == "add" {
		var dev *backive.Device
		var uuid string
		fsUUID, ok := envMap["ID_FS_UUID"]
		if !ok {
			log.Println("ID_FS_UUID not available ?!")
			return
		}
		uuid = fsUUID
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
			backups, found := config.Backups.FindBackupsForDevice(*dev)
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
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChan
			switch s {
			case syscall.SIGHUP:
				log.Println("hungup")
			case syscall.SIGINT:
				log.Println("Ctrl+C, quitting.")
				exitChan <- 0
			case syscall.SIGTERM:
				log.Println("Terminating.")
				exitChan <- 0
			case syscall.SIGQUIT:
				log.Println("Quitting")
				exitChan <- 0
			default:
				log.Println("Unknown signal.")
				exitChan <- 1
			}
		}
	}()
	go func() {
		// exit function only does something when the exit_chan has an item
		// cleaning up stuff
		code := <-exitChan
		database.Save()
		log.Printf("Received exit code (%d), shutting down.", code)
		os.Exit(code)
	}()

	// TODO: do proper signal handling!
	log.Println("backive starting up...")
	// find and load config
	database.Load()
	config.Load()
	setupLogging()
	backive.Init(config, database)

	// init scheduler and check for next needed runs?
	// start event loop
	events.Init(config.Settings.UnixSocketLocation)
	events.RegisterCallback(defaultCallback)
	// accept events
	events.Listen()
}
