package events

import "net"

var ls net.Listener
var done <-chan struct{}

func Init(socketPath string) {
	ls, err = net.Listen(socketPath)
	if err != nil {
		panic(err)
	}
}

func RunLoop() {
	for {
		go func() {
			process()
		}()
	}

}

func process() {
	client, err = ls.Accept()
	if err != nil {
		panic(err)
	}
	//TODO: rewrite to be safe regarding buffer length
	buf := make([]byte, 2048)
	nr, err := client.Read(buf)
	if err != nil {
		return
	}

	data := buf[0:nr]
}
