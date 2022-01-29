package backive

import (
	"encoding/json"
	"log"
	"os"
	"path"
)

var mockOsWriteFile = os.WriteFile
var mockOsReadFile = os.ReadFile

// Database is a simple string to string mapping, where arbitrary strings can be stored and safed to disk or loaded
type Database struct {
	data map[string]string
}

var dbPath = "/var/lib/backive/data.json"

// Save saves the database
func (d *Database) Save() {
	jsonstr, merr := json.Marshal(d.data)
	if merr != nil {
		panic(merr)
	}
	log.Printf("Writing database output to file: %s", jsonstr)
	saveDir, _ := path.Split(dbPath)
	CreateDirectoryIfNotExists(saveDir)
	err := mockOsWriteFile(dbPath, []byte(jsonstr), 0644)
	if err != nil {
		panic(err)
	}
}

// Load loads the database
func (d *Database) Load() {
	if _, err := os.Stat(dbPath); err == nil {
		data, rferr := mockOsReadFile(dbPath)
		if rferr != nil {
			panic(rferr)
		}
		json.Unmarshal(data, &d.data)
	} /*else if os.IsNotExist(err) {
		// no data
	}*/
}
