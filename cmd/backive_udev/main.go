package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// main Simple main function for the udev callback executable, registered with the udev service.
func main() {
	udevLogdir := "/var/log/backive"
	udevLogname := "/var/log/backive/udev.log"
	if _, err := os.Stat(udevLogdir); err == nil {
		//ignore
	} else if os.IsNotExist(err) {
		os.MkdirAll(udevLogdir, 0755)
	}
	f, err := os.OpenFile(udevLogname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Error creating logfile!")
		panic("no logfile no info")
	}
	defer f.Close()

	log.SetOutput(f)

	env := map[string]string{}

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		env[pair[0]] = pair[1]
		log.Println(e)
	}

	c, err := net.Dial("unix", "/var/local/backive/backive.sock")
	if err != nil {
		log.Fatalln("Could not instantiate unix socket. Aborting")
	}
	jsonstr, err := json.Marshal(env)
	if err != nil {
		log.Fatalln("Could not convert to json. Aborting")
	}
	c.Write(jsonstr)
	defer c.Close()
	log.Printf("Sent %d bytes to unix socket.\n", len(jsonstr))
}
