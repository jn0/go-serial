package sio

import "fmt"
import "strings"
import "testing"

const USB0 = "/dev/ttyUSB0"

func TestMain(t *testing.T) {
	var p Port

	fmt.Println("begin")

	e := p.Open(USB0)
	if e != nil { panic(e); }
	defer p.Close()
	fmt.Println("Port is open")

	tty, _ := NewConsole(&p); defer tty.Close()

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
