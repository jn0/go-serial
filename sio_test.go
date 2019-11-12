package sio

import "fmt"
import "strings"
import "testing"

import (
	"os"
	"os/signal"
	"syscall"
)

const USB0 = "/dev/ttyUSB0"

var tty *Console
var port *Port

func sighand(sigchan chan os.Signal) {
	for {
		fmt.Println("Waiting for signal")
		sig := <- sigchan // wait for a signal
		if tty != nil { tty.Close(); fmt.Println("tty restored"); }
		panic(fmt.Sprintf("Killed with: %v", sig))
	}
}

func userEnd(end chan bool, com chan string) {
	var s, resp string
	fmt.Println("User start")
	var i int = 0
	for {
		tty.Write(fmt.Sprintf("[%d]\n", i)) ; i += 1
		s, _ = tty.ReadLine()		// tty -> (s)
		if strings.HasPrefix(s, "quit") {
			break
		}
		com <- s
		select { case resp = <- com: tty.Write(resp); }
	}
	end <- true
	fmt.Println("User end")
}

func modemEnd(end chan bool, com chan string) {
	var cmd, resp string
	fmt.Println("Modem start")
	for {
		select {
		case cmd = <- com:
			tty.Send(cmd)			// (s) -> usb
			resp, _ = tty.RecvUntil(Stops)	// usb -> (s)
			com <- resp
		case <- end:
			break
		}
	}
	fmt.Println("Modem end")
}

func TestMain(t *testing.T) {
	fmt.Println("begin")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go sighand(sig)

	port = NewSerialPort(USB0)
	fmt.Println("Port is open:", port.String())

	maj, min := port.DeviceId()
	fmt.Printf("[%d:%d] %s\n", maj, min, port.DeviceClassName())

	fmt.Println("Located:")
	PrintLocations(port.SysFS(), fmt.Printf)


	tty, _ = NewConsole(port)
	defer func() { tty.Close(); fmt.Println("tty restored"); }()

	end := make(chan bool)		// termination mark
	com := make(chan string)	// data xchg

	go userEnd(end, com)
	go modemEnd(end, com)

	<-end

	fmt.Println("end")
}

/* EOF */
