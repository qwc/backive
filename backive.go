package backive

var config Configuration
var runs Runs
var database Database

// Init initializes backive with the two basic data structures required, the config, and the database
func Init(cfg Configuration, db Database) {
	config = cfg
	database = db
	runs.Load(database)
}
