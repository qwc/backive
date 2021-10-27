package db

import (
	"encoding/json"
	"os"
)

var Database map[string]string
var path string = "/var/lib/backive/data.json"

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

func Load() {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, Database)
}
