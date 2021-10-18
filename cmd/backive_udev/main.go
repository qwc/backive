package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	f, err := os.OpenFile("/tmp/backive/udev.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
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
		log.Println("%s", e)
	}

	c, err := net.Dial("unix", "/tmp/backive/backive.sock")
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
