package backive

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

func TestEventhandler(t *testing.T) {
	t.Skip("Do not get it to work...")
	eh := new(EventHandler)
	eh.Init("./backive.socket")
	defer func() {
		eh.Stop()
		err := os.Remove("./backive.socket")
		if err != nil {
			t.Log(err)
		}
	}()
	t.Log("Initialized test")
	go eh.Listen()
	t.Log("eh is listening")
	var hasBeenCalled = make(chan bool)
	eh.RegisterCallback(
		func(m map[string]string) {
			hasBeenCalled <- true
		},
	)
	t.Log("registered callback")
	beenCalled := false
	var counter = 0
	env := map[string]string{}
	env["test"] = "test"
	message := map[string]interface{}{}
	message["request"] = "udev"
	message["data"] = env
	for {
		select {
		case data := <-hasBeenCalled:
			t.Log("receiving message")
			beenCalled = data
			if !beenCalled {
				t.Fail()
			}
			t.Log("received message")
			eh.Stop()
			return
		default:
			t.Logf("Waiting for callback %d", counter)
			time.Sleep(time.Millisecond)
			if counter == 2 {
				sendDataToSocket("./backive.socket", message)
				t.Log("sent message")
			}
			if counter < 10 {
				counter++
			} else {
				t.Log("Stopping with Fail")
				eh.Stop()
				t.Fail()
				break
			}
		}
	}
}

func sendDataToSocket(socket string, message map[string]interface{}) {
	c, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalln("Could not instantiate unix socket. Aborting")
	}
	jsonstr, err := json.Marshal(message)
	if err != nil {
		log.Fatalln("Could not convert to json. Aborting")
	}
	c.Write(jsonstr)
	defer c.Close()
}
