package backive

import (
	"log"
	"os"
)

var mockOsStat = os.Stat
var mockOsMkdirAll = os.MkdirAll

// CreateDirectoryIfNotExists Checks for a directory string and creates the directory if it does not exist, must be a absolute path.
func CreateDirectoryIfNotExists(dir string) {
	if _, err := mockOsStat(dir); err == nil {
		//ignore
	} else if os.IsNotExist(err) {
		mockOsMkdirAll(dir, 0755)
	} else {
		log.Fatal(err)
	}
}
