package backive

import (
	"log"
	"os"
)

// CreateDirectoryIfNotExists Checks for a directory string and creates the directory if it does not exist, must be a absolute path.
func CreateDirectoryIfNotExists(dir string) {
	if _, err := os.Stat(dir); err == nil {
		//ignore
	} else if os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	} else {
		log.Fatal(err)
	}
}
