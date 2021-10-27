package scheduler

import (
	"encoding/json"
	"time"

	"github.com/qwc/backive/config"
	"github.com/qwc/backive/db"
)

// Runs contains the Data for the scheduler: mapping from backups to a list of timestamps of the last 10 backups
type Runs map[string][]time.Time

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
	if freq > 0 {
		return true
	}

	return false
}
