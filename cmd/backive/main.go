package main

import (
	"fmt"

	"github.com/qwc/backive/config"
)

func main() {
	// TODO: do proper signal handling!
	fmt.Println("vim-go")
	// find and load config
	config.Load()

	// init scheduler and check for next needed runs?

	// start event loop
	// accept event
	// find associated device and it's backups
	// mount device
	// run backups, one after another if multiple
	// unmount device
	// end loop

	// cleanup if anything is there to cleanup
}
