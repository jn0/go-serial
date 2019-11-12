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

func printf(format string, args ...interface{}) (int, error) {
	fmt.Fprintf(os.Stderr, format + "\r\n", args...)
	return 0, nil
}

func sighand(sigchan chan os.Signal) {
	for {
		printf("Waiting for signal")
		sig := <- sigchan // wait for a signal
		if tty != nil { tty.Close(); printf("tty restored"); }
		panic(fmt.Sprintf("Killed with: %v", sig))
	}
}

func userEnd(end chan bool, com chan []byte) {
	printf("User start")
	for {
		if s, e := tty.ReadLine(); e != nil {
			panic(e)
		} else if strings.HasPrefix(s, "quit") {
			break
		} else {
			com <- []byte(s + "\r")
		}
	}
	end <- true
	printf("User end")
}

func modemEnd(end chan bool, com chan []byte) {
	printf("Modem start")
	for port.IsOpen() {
		select {
		case cmd := <- com: port.Write(cmd)
		case <- end: printf("\r\nmodem QUIT"); break
		default:
			if n, e := port.InWaiting(); e != nil {
				if e == PortNotOpenError {
					printf("\r\nport lost")
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
	printf("Modem end")
}

func TestMain(t *testing.T) {
	printf("begin")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go sighand(sig)

	port = NewSerialPort(USB0)
	defer func() { port.Close(); printf("\r\nport closed"); }()
	printf("Port is open: %s", port)

	printf("Sys FS:"); PrintLocations(port.SysFS(), printf)

	tty, _ = NewConsole()
	defer func() { tty.Close(); printf("\r\ntty restored"); }()
	tty.Write([]byte("\r\n"))

	end := make(chan bool)		// termination mark
	com := make(chan []byte)	// data xchg

	go userEnd(end, com)
	go modemEnd(end, com)

	<-end

	printf("the end")
}

/* EOF */
