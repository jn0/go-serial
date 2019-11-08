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

func sighand(sigchan chan os.Signal) {
	for {
		fmt.Println("Waiting for signal")
		sig := <- sigchan // wait for a signal
		if tty != nil { tty.Close(); fmt.Println("tty restored"); }
		panic(fmt.Sprintf("Killed with: %v", sig))
	}
}

func TestMain(t *testing.T) {
	var p Port

	fmt.Println("begin")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go sighand(sig)

	e := p.Open(USB0)
	if e != nil { panic(e); }
	fmt.Println("Port is open:", p.String())

	tty, _ = NewConsole(&p); defer func() { tty.Close(); fmt.Println("tty restored"); }()

	var i int = 0
	for {
		tty.Write(fmt.Sprintf("[%d]\n", i)) ; i += 1
		s, _ := tty.ReadLine()		// tty -> (s)
		if strings.HasPrefix(s, "quit") {
			break
		}
		tty.Send(s)			// (s) -> usb
		s, _ = tty.RecvUntil(Stops)	// usb -> (s)
		tty.Write(s)			// (s) -> tty
	}

	fmt.Println("end")
}

/* EOF */
