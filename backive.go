package backive

var config Configuration
var runs Runs
var database Database

func Init(cfg Configuration, db Database) {
	config = cfg
	database = db
	runs.Load(database)
}
