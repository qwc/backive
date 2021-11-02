package scheduler

import (
	"container/list"
	"encoding/json"
	"fmt"
	"time"

	"github.com/qwc/backive/config"
	"github.com/qwc/backive/db"
)

type backupRuns struct {
	runlist *list.List
}

// Runs contains the Data for the scheduler: mapping from backups to a list of timestamps of the last 10 backups
type Runs map[string]backupRuns

var runs Runs

// Load loads the data from the json database
func Load() {
	runerr := json.Unmarshal([]byte(db.Database["runs"]), &runs)
	if runerr != nil {
		panic(runerr)
	}
}

// Save saves the data into the json database
func Save() {
	str, err := json.Marshal(runs)
	if err != nil {
		panic(err)
	}
	db.Database["runs"] = string(str)
}

// ShouldRun Takes a backup key and returns a bool if a backup should run now.
func ShouldRun(backup string) bool {
	backupdata := config.Get().Backups[backup]
	freq := backupdata.Frequency
	// calculate time difference from last run, return true if no run has taken place
	lr, ok := LastRun(backup)
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
func RegisterRun(backup string) {
	nbl, ok := runs[backup]
	if !ok {
		nbl.runlist = list.New()
		runs[backup] = nbl
	}
	nbl.runlist.PushFront(time.Now())
}

// LastRun returns the time.Time of the last run of the backup given.
func LastRun(backup string) (time.Time, error) {
	_, ok := runs[backup]
	if ok {
		var t = time.Time(runs[backup].runlist.Front().Value.(time.Time))
		return t, nil
	}
	return time.Unix(0, 0), fmt.Errorf("Backup name not found and therefore has never run")
}
