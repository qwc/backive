package backive

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
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

type MsgLvls struct {
	Error  MsgLevel
	Finish MsgLevel
	Remind MsgLevel
	Info   MsgLevel
	Debug  MsgLevel
}

// UIHandler internal data struct
type UIHandler struct {
	ls     net.Listener
	client net.Conn
}

var UiHdl UIHandler
var MsgLevels = MsgLvls{ERROR, FINISH, REMIND, INFO, DEBUG}

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
	os.Chmod(socketPath, 0777)
	if err != nil {
		log.Printf("Error: %s", err)
		return err
	}
	log.Println("Listening for ui clients")
	return nil
}

// Listen starts the Unix socket listener
func (uh *UIHandler) Listen() {
	log.Println("Running UIHandler loop")
	for {
		var err error
		uh.client, err = uh.ls.Accept()
		if uh.client == nil {
			log.Println("Client is nil, why?")
		}
		if uh.client != nil {
			log.Printf("client's local addr %s", uh.client.LocalAddr().String())
		}
		if err != nil {
			log.Printf("Accept failed %e\n", err)
		}
		log.Print("Accepted UI client")
	}
}

// DisplayMessage is the method to use inside the service to display messages, with intended level
func (uh *UIHandler) DisplayMessage(header string, message string, level MsgLevel) error {
	if uh.client != nil {
		var data = make(map[string]string)
		data["level"] = fmt.Sprint(level)
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
