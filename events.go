package backive

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"path"
)

var mockAccept = func(eh *EventHandler) (net.Conn, error) {
	log.Println("Calling eh.ls.Accept()")
	return eh.ls.Accept()
}

// EventHandler holds the necessary elements to get an eventhandler setup and working.
type EventHandler struct {
	ls        net.Listener
	callbacks []func(map[string]string)
}

// Init initializes the unix socket.
func (eh *EventHandler) Init(socketPath string) error {
	log.Println("Initializing EventHandler...")
	var err error
	dir, _ := path.Split(socketPath)
	CreateDirectoryIfNotExists(dir)
	eh.ls, err = net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	eh.callbacks = make([]func(map[string]string), 3)
	return nil
}

// Listen starts the event loop.
func (eh *EventHandler) Listen() {
	log.Println("Running eventloop")
	func() {
		for {
			eh.process()
		}
	}()
}

// RegisterCallback adds a function to the list of callback functions for processing of events.
func (eh *EventHandler) RegisterCallback(cb func(map[string]string)) {
	eh.callbacks = append(eh.callbacks, cb)
}

// process processes each and every unix socket event, Unmarshals the json data and calls the list of callbacks.
func (eh *EventHandler) process() {
	client, err := mockAccept(eh)
	log.Println("Accepted client")
	UiHdl.DisplayMessage("Event debugging", "Catched event...", MsgLevels.Debug)
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()
	data := make([]byte, 2048)
	for {
		buf := make([]byte, 512)
		nr, err := client.Read(buf)
		log.Printf("Read %d bytes...", nr)
		if err != nil && err != io.EOF {
			log.Println(err)
			return
		}
		data = append(data, buf[0:nr]...)
		if err == io.EOF {
			break
		}
	}
	sdata := string(bytes.Trim(data, "\x00"))
	var message map[string]interface{}
	log.Printf("Reading JSON: %s", sdata)
	errjson := json.Unmarshal([]byte(sdata), &message)
	if errjson != nil {
		log.Println(errjson)
		return
	}
	log.Println("Calling callbacks")
	var env = map[string]string{}
	if message["request"] == "udev" {
		for k, v := range message["data"].(map[string]interface{}) {
			env[k] = v.(string)
		}
		UiHdl.DisplayMessage("Event debugging", "Got udev event msg.", MsgLevels.Debug)
	}
	for _, v := range eh.callbacks {
		if v != nil {
			log.Println("Calling callback")
			v(env)
		}
	}
}
