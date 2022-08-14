package backive

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"path"
)

// MsgLevel type for setting the message level
type MsgLevel int

// Constants for setting the MsgLevel
const (
	ERROR MsgLevel = iota * 10
	FINISH
	REMIND
	INFO
	DEBUG
)

// UIHandler internal data struct
type UIHandler struct {
	ls     net.Listener
	client net.Conn
}

var mockUIAccept = func(uh *UIHandler) (net.Conn, error) {
	log.Println("Calling eh.ls.Accept()")
	return uh.ls.Accept()
}

// Init initializing the UIHandler
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

// Listen starts the Unix socket listener
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

// DisplayMessage is the method to use inside the service to display messages, with intended level
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
	}
	log.Println("No UI client available, msg did not get delivered.")
	return fmt.Errorf("No UI client available, msg not delivered")
}
