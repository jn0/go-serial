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
		fmt.Println("Waiting for signal\r")
		sig := <- sigchan // wait for a signal
		if tty != nil { tty.Close(); fmt.Println("tty restored"); }
		panic(fmt.Sprintf("Killed with: %v", sig))
	}
}

func userEnd(end chan bool, com chan []byte) {
	fmt.Println("User start\r")
	for {
		select {
		case <- end: tty.Write([]byte("\r\nuser QUIT\r\n")); break
		default:
			if s, e := tty.ReadLine(); e != nil {
				panic(e)
			} else {
				if strings.HasPrefix(s, "quit") {
					end <- true
					break
				}
				com <- []byte(s + "\r")
			}
		}
	}
	fmt.Println("User end\r")
}

func modemEnd(end chan bool, com chan []byte) {
	fmt.Println("Modem start\r")
	for port.IsOpen() {
		select {
		case cmd := <- com: port.Write(cmd)
		case <- end: tty.Write([]byte("\r\nmodem QUIT\r\n")); break
		default:
			if n, e := port.InWaiting(); e != nil {
				if e == PortNotOpenError {
					break
				}
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
	fmt.Println("Modem end\r")
}

func TestMain(t *testing.T) {
	fmt.Println("begin")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go sighand(sig)

	port = NewSerialPort(USB0)
	defer func() { port.Close(); fmt.Println("\r\nport closed\r"); }()
	fmt.Println("Port is open:", port.String())

	maj, min := port.DeviceId()
	fmt.Printf("[%d:%d] %s\n", maj, min, port.DeviceClassName())

	fmt.Println("Located:")
	PrintLocations(port.SysFS(), fmt.Printf)

	tty, _ = NewConsole()
	defer func() { tty.Close(); fmt.Println("\r\ntty restored\r"); }()
	tty.Write([]byte("\r\n"))

	end := make(chan bool)		// termination mark
	com := make(chan []byte)	// data xchg

	go userEnd(end, com)
	go modemEnd(end, com)

	<-end

	fmt.Println("end")
}

/* EOF */
