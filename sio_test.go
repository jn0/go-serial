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

func closemall() {
	if tty != nil { tty.Close(); printf("signal: tty restored"); }
	if port != nil { port.Close(); printf("signal: port closed"); }
}

func sighand(sigchan chan os.Signal) {
	for {
		printf("Waiting for signal")
		sig := <- sigchan // wait for a signal
		closemall()
		panic(fmt.Sprintf("Killed with: %v", sig))
	}
}

type StopChannel chan bool

func (self *StopChannel) Wait() {
	<- *self
}
func (self *StopChannel) Notify() {
	*self <- true
}

func userEnd(end StopChannel, com CommandChannel) {
	tty.WriteString("\r\nType 'quit' to terminate.\r\n")
	for {
		if s, e := tty.ReadLine(); e != nil {
			panic(e)
		} else if strings.HasPrefix(s, "quit") {
			break
		} else {
			com.SendString(s + "\r")
		}
	}
	tty.WriteString("\r\nTerminating...\r\n")
	end.Notify()
}

func TestMain(t *testing.T) {
	printf("begin")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go sighand(sig)

	port = NewSerialPort(USB0)
	defer func() { port.Close(); printf("port closed"); }()
	printf("Port is open: %s", port)

	printf("Sys FS:"); PrintLocations(port.SysFS(), printf)

	tty, _ = NewConsole()
	defer func() { tty.Close(); printf("\r\ntty restored"); }()
	tty.Write([]byte("\r\n"))

	end := make(StopChannel)
	cmd := make(CommandChannel)

	go userEnd(end, cmd)
	go port.Interact(cmd, tty.Write)

	cmd.SendString("ATI\r")

	end.Wait()

	printf("the end")
}

/* EOF */
