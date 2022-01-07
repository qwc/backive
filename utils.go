package backive

import (
	"log"
	"os"
)

func CreateDirectoryIfNotExists(dir string) {
	if _, err := os.Stat(dir); err == nil {
		//ignore
	} else if os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	} else {
		log.Fatal(err)
	}
}
