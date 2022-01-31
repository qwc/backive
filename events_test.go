package backive

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

type ConnStub struct {
}

var counter int

func (c ConnStub) Read(b []byte) (int, error) {
	switch {
	case counter == 0:
		counter++
		env := map[string]string{}
		env["test"] = "test"
		message := map[string]interface{}{}
		message["request"] = "udev"
		message["data"] = env
		data, err := json.Marshal(message)
		copy(b, data)
		log.Println(string(b))
		return len(data), err
	case counter == 1:
		counter++
		return 0, io.EOF
	case counter == 2:
		counter++
		return 0, fmt.Errorf("Some Error for testing")
	default:
		return 0, io.EOF
	}
}
func (c ConnStub) Close() error {
	return nil
}
func (c ConnStub) LocalAddr() net.Addr {
	return nil
}
func (c ConnStub) RemoteAddr() net.Addr {
	return nil
}
func (c ConnStub) SetDeadline(t time.Time) error {
	return nil
}
func (c ConnStub) SetReadDeadline(t time.Time) error {
	return nil
}
func (c ConnStub) SetWriteDeadline(t time.Time) error {
	return nil
}
func (c ConnStub) Write(b []byte) (int, error) {
	return 0, nil
}

var hasBeenCalled = false

func TestEventhandler(t *testing.T) {
	eh := new(EventHandler)
	err := eh.Init("./backive.socket")
	if err != nil {
		t.Fail()
	}
	defer func() {
		err := os.Remove("./backive.socket")
		if err != nil {
			t.Log(err)
		}
	}()
	t.Log("Initialized test")
	//var hasBeenCalled = make(chan bool)
	eh.RegisterCallback(
		func(m map[string]string) {
			t.Log("Callback got called")
			hasBeenCalled = true
		},
	)
	t.Log("registered callback")

	mockAccept = func(eh *EventHandler) (net.Conn, error) {
		t.Log("Mocked Accept() has been called.")
		mycon := ConnStub{}
		return mycon, nil
	}
	eh.process()
	//beenCalled := <-hasBeenCalled
	if !hasBeenCalled {
		t.Log("Got false, need true.")
		t.Fail()
	}
}
