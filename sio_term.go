package sio

import "fmt"
import "os"
import "golang.org/x/crypto/ssh/terminal"

const TTY = "/dev/tty"
const Prompt = "> "

type Console struct {
	port *Port
	tty *os.File
	ttyFd int
	con *terminal.Terminal
	state *terminal.State
}
func (self *Console) Close() (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	terminal.Restore(self.ttyFd, self.state)
	self.ttyFd = -1
	self.tty.Close()
	if self.port != nil && self.port.IsOpen() {
		self.port.Close()
	}
	return nil
}
func (self *Console) Open(p *Port) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	self.port = p

	self.tty, e = os.OpenFile(TTY, os.O_RDWR, 0)
	assert(e, "Open(%+q): %w", TTY, e)
	self.ttyFd = int(self.tty.Fd())

	if !terminal.IsTerminal(self.ttyFd) {
		assert(fmt.Errorf("stdin is not a terminal"), "Open")
	}

	self.state, e = terminal.MakeRaw(self.ttyFd)
	assert(e, "MakeRaw")

	self.con = terminal.NewTerminal(self.tty, Prompt)
	return nil
}
// RecvUntil() reads from serial port
func (self *Console) RecvUntil(stops []string) (s string, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	s, e = self.port.ReadUntil(stops)
	assert(e, "RecvUntil")
	return s, nil
}
// Send() writes to the serial port
func (self *Console) Send(s string) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	self.port.WriteLine(s)
	return nil
}
// Write() writes to user's screen
func (self *Console) Write(s string) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	self.con.Write([]byte(s))
	return nil
}
// ReadLine() reads a line from user's keyboard
func (self *Console) ReadLine() (s string, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	s, e = self.con.ReadLine()
	assert(e, "ReadLine")
	return s, nil
}

func NewConsole(p *Port) (con *Console, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	con = &Console{}
	assert(con.Open(p), fmt.Sprintf("Console.Open(%+v)", p))

	return con, nil
}

/* EOF */
