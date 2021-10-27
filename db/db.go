package db

import (
	"encoding/json"
	"os"
)

// Database is a simple string to string mapping, where arbitrary strings can be stored and safed to disk or loaded
var Database map[string]string
var path string = "/var/lib/backive/data.json"

// Save saves the database
func Save() {
	jsonstr, merr := json.Marshal(Database)
	if merr != nil {
		panic(merr)
	}

	err := os.WriteFile(path, []byte(jsonstr), 0644)
	if err != nil {
		panic(err)
	}
}

// Load loads the database
func Load() {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &Database)
}
