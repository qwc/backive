package db

import "encoding/json"

var database map[string]string

func Save() {
	jsonstr := json.Marshal(database)

}
