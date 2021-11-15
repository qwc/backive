package events

import (
	"encoding/json"
	"io"
	"net"
)

var ls net.Listener
var done <-chan struct{}
var callbacks []func(map[string]string)

// Init initializes the unix socket.
func Init(socketPath string) {
	var err error
	ls, err = net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	callbacks = make([]func(map[string]string), 3)
}

// Listen starts the event loop.
func Listen() {
	for {
		go func() {
			process()
		}()
	}
}

// RegisterCallback adds a function to the list of callback functions for processing of events.
func RegisterCallback(cb func(map[string]string)) {
	callbacks = append(callbacks, cb)
}

// process processes each and every unix socket event, Unmarshals the json data and calls the list of callbacks.
func process() {
	client, err := ls.Accept()
	if err != nil {
		panic(err)
	}
	data := make([]byte, 2048)
	for {
		buf := make([]byte, 512)
		nr, err := client.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		data = append(data, buf[0:nr]...)
		if err == io.EOF {
			break
		}
	}
	env := map[string]string{}
	errjson := json.Unmarshal(data, &env)
	if errjson != nil {
		panic(errjson)
	}
	for _, v := range callbacks {
		v(env)
	}
}
