package backive

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"path"
)

type UIHandler struct {
	ls     net.Listener
	client net.Conn
}

var mockUIAccept = func(uh *UIHandler) (net.Conn, error) {
	log.Println("Calling eh.ls.Accept()")
	return uh.ls.Accept()
}

func (uh *UIHandler) Init(socketPath string) error {
	log.Println("Initializing UIHandler")
	var err error
	dir, _ := path.Split(socketPath)
	CreateDirectoryIfNotExists(dir)
	uh.ls, err = net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	return nil
}

func (uh *UIHandler) Listen() {
	log.Println("Running UIHandler loop")
	func() {
		for {
			var err error
			uh.client, err = uh.ls.Accept()
			if err != nil {
				log.Printf("Accept failed %e\n", err)
			}
		}
	}()
}

func (uh *UIHandler) DisplayMessage(header string, message string, level int) error {
	if uh.client != nil {
		var data map[string]interface{}
		data["level"] = level
		data["header"] = header
		data["message"] = message
		b, err := json.Marshal(data)
		if err != nil {
			log.Printf("Problem in sending message to UI: %e", err)
			return err
		}
		uh.client.Write(b)
		return nil
	} else {
		log.Println("No UI client available, msg did not get delivered.")
		return fmt.Errorf("No UI client available, msg not delivered.")
	}
}
