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

func userEnd(end chan bool, com chan []byte) {
	fmt.Println("User start")
	for {
		if s, e := tty.ReadLine(); e != nil {
			panic(e)
		} else {
			if strings.HasPrefix(s, "quit") {
				break
			}
			com <- []byte(s + "\r")
		}
	}
	end <- true
	fmt.Println("User end")
}

func modemEnd(end chan bool, com chan []byte) {
	fmt.Println("Modem start")
	for port.IsOpen() {
		select {
		case cmd := <- com: port.Write(cmd)
		case <- end: break
		default:
			if n, e := port.InWaiting(); e != nil {
				panic(e)
			} else if n > 0 {
				b := make([]byte, n)
				if n, e := port.Read(b); e != nil {
					panic(e)
				} else if n == 0 {
					continue
				}
				tty.Write(b)
			}
		}
	}
	fmt.Println("Modem end")
}

func TestMain(t *testing.T) {
	fmt.Println("begin")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go sighand(sig)

	port = NewSerialPort(USB0); defer port.Close()
	fmt.Println("Port is open:", port.String())

	maj, min := port.DeviceId()
	fmt.Printf("[%d:%d] %s\n", maj, min, port.DeviceClassName())

	fmt.Println("Located:")
	PrintLocations(port.SysFS(), fmt.Printf)

	tty, _ = NewConsole()
	defer func() { tty.Close(); fmt.Println("tty restored"); }()

	end := make(chan bool)		// termination mark
	com := make(chan []byte)	// data xchg

	go userEnd(end, com)
	go modemEnd(end, com)

	<-end

	fmt.Println("end")
}

/* EOF */
