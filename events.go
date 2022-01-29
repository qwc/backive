package backive

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"path"
)

// EventHandler holds the necessary elements to get an eventhandler setup and working.
type EventHandler struct {
	ls net.Listener
	//done      <-chan struct{}
	callbacks []func(map[string]string)
	stop      chan bool
}

// Init initializes the unix socket.
func (eh *EventHandler) Init(socketPath string) {
	eh.stop = make(chan bool)
	log.Println("Initializing EventHandler...")
	var err error
	dir, _ := path.Split(socketPath)
	CreateDirectoryIfNotExists(dir)
	eh.ls, err = net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	eh.callbacks = make([]func(map[string]string), 3)
}

// Stop stops the Eventhandler
func (eh *EventHandler) Stop() {
	log.Println("Closing EventHandler")
	eh.stop <- true
	err := eh.ls.Close()
	if err != nil {
		log.Println("Error closing the listener")
	}
	log.Println("Closed EventHandler")
}

// Listen starts the event loop.
func (eh *EventHandler) Listen() {
	log.Println("Running eventloop")
	func() {
		for {
			select {
			case <-eh.stop:
				return
			default:
				eh.process()
			}
		}
	}()
}

// RegisterCallback adds a function to the list of callback functions for processing of events.
func (eh *EventHandler) RegisterCallback(cb func(map[string]string)) {
	eh.callbacks = append(eh.callbacks, cb)
}

// process processes each and every unix socket event, Unmarshals the json data and calls the list of callbacks.
func (eh *EventHandler) process() {
	client, err := eh.ls.Accept()
	log.Println("Accepted client")
	if err != nil {
		select {
		case <-eh.stop:
			return
		default:
			log.Fatal(err)
		}
	}
	defer client.Close()
	data := make([]byte, 2048)
	for {
		buf := make([]byte, 512)
		nr, err := client.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		data = append(data, buf[0:nr]...)
		if err == io.EOF {
			break
		}
	}
	sdata := string(bytes.Trim(data, "\x00"))
	//log.Println(sdata)
	var message map[string]interface{}
	errjson := json.Unmarshal([]byte(sdata), &message)
	if errjson != nil {
		log.Fatal(errjson)
	}
	var env = map[string]string{}
	if message["request"] == "udev" {
		for k, v := range message["data"].(map[string]interface{}) {
			env[k] = v.(string)
		}
	}
	for _, v := range eh.callbacks {
		if v != nil {
			v(env)
		}
	}
}
