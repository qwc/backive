package scheduler

import (
	"encoding/json"
	"time"

	"github.com/qwc/backive/config"
	"github.com/qwc/backive/db"
)

type Runs map[string][]time.Time

var runs Runs

func Load() {
	runerr := json.Unmarshal([]byte(db.Database["runs"]), &runs)
	if runerr != nil {
		panic(runerr)
	}
}

func Save() {
	str, err := json.Marshal(runs)
	if err != nil {
		panic(err)
	}

	db.Database["runs"] = string(str)
}

func ShouldRun(backup string) bool {
	freq := config.Get().Backups[backup].Frequency

	// calculate time difference from last run, return true if no run has taken place
	if freq > 0 {
		return true
	}

	return false
}
