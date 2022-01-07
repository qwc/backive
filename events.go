package backive

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"path"
)

type EventHandler struct {
	ls        net.Listener
	done      <-chan struct{}
	callbacks []func(map[string]string)
}

// Init initializes the unix socket.
func (eh *EventHandler) Init(socketPath string) {
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
	client, err := eh.ls.Accept()
	log.Println("Accepted client")
	if err != nil {
		log.Fatal(err)
	}
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
	env := map[string]string{}
	errjson := json.Unmarshal([]byte(sdata), &env)
	if errjson != nil {
		log.Fatal(errjson)
	}
	for _, v := range eh.callbacks {
		if v != nil {
			v(env)
		}
	}
}
