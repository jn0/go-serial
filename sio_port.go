package sio

import (
	"os"
	"io"
	"time"
	"strings"
	"syscall"
	"runtime"
	"sync"
	"fmt"
)

const PortOpenFlags = os.O_RDWR | syscall.O_NOCTTY | syscall.O_NONBLOCK
const DefaultTimeout = time.Millisecond * 50 // 0.05 sec
const CharacterSpecialDeviceMode = os.ModeDevice | os.ModeCharDevice

func exists(path string) bool {
	_, e := os.Stat(path)
	return e == nil || !os.IsNotExist(e)
}

func isCharDevice(stat os.FileInfo) bool {
	return stat.Mode() & CharacterSpecialDeviceMode == CharacterSpecialDeviceMode
}

type Port struct {
	file *os.File
	stat os.FileInfo
	fd Ioctl
	exclusive bool
	inter_byte_timeout float64
	speed BitRate
	char_size CharSize
	parity Parity
	stop_bits StopBits
	xonxoff, rtscts, dsrdtr bool
	rs485 Rs485
	termios syscall.Termios
	pipe struct {
		abort_read, abort_write servicePipe
	}
	sysfs []string
	lock sync.Mutex
}

// NewSerialPort("/dev/ttyUSB0") returns ref to an open Port instance
func NewSerialPort(dev string) *Port {
	var p *Port = &Port{}
	e := p.Open(dev)
	assert(e, "NewSerialPort(%+q): %w", dev, e)
	return p
}

func (self *Port) SysFS() []string {
	return self.sysfs
}

func (self *Port) String() string {
	if self.IsOpen() {
		major, minor := self.DeviceId()
		return fmt.Sprintf("<sio.Port(%+q):%s [%d:%d]>",
					self.file.Name(),
					self.DeviceClassName(), major, minor)
	} else {
		return "<sio.Port>"
	}
}
func (self *Port) IsOpen() bool {
	return self.file != nil && self.file.Name() != ""
}
func (self *Port) Open(path string) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	self.lock.Lock(); defer self.lock.Unlock()

	stat, e := os.Stat(path)
	if e != nil {
		if os.IsNotExist(e) {
			assert(e, "Port.Open(%+q): no path", path)
		}
		assert(e, "Port.Open(%+q): %v", path, e)
	}
	assertb(isCharDevice(stat), "Port.Open(%+q): not a character device", path)
	assertb(!self.IsOpen(), "Port.Open(%+q): *object* is already open", path)

	self.stat = stat

	self.exclusive = true
	self.speed = BIT_RATE_B9600

	// "8N1"
	self.char_size, self.parity, self.stop_bits = CHAR_SIZE_8, PARITY_NONE, STOP_BITS_1

	self.xonxoff = false
	self.rtscts = false
	self.dsrdtr = false
	self.rs485.Enabled = false

	self.file, e = os.OpenFile(path, PortOpenFlags, 0)
	assert(e, "Port.Open(%+q)@OpenFile: %w", path, e)

	self.fd.Set(self.file.Fd())

	major, minor := self.DeviceId()
	var sysfs SysFS
	sysfs.Use(GetRDev(self.stat))
	self.sysfs = sysfs.Locate(SysfsClass, major, minor)

	e = self.fd.Reconfigure(&self.termios, self)
	assert(e, "Open: cannot configure port")

	if !self.dsrdtr {
		self.fd.SetDTR(false) // doesn't work actually
		// assert(self.fd.SetDTR(false), "SetDTR")
	}
	if !self.rtscts {
		self.fd.SetRTS(false) // doesn't work actually
		// assert(self.fd.SetRTS(false), "SetRTS")
	}
	assert(self.ResetInput(), "ResetInput")
	assert(self.ResetOutput(), "ResetOutput")
	assert(self.pipe.abort_read.Open(), "pipe(read)")
	assert(self.pipe.abort_write.Open(), "pipe(write)")

	return nil
}
func (self *Port) Close() error {
	self.lock.Lock(); defer self.lock.Unlock()

	if self.file != nil {
		self.file.Close()
		self.file = nil
		self.fd = ZeroIoctl
		for _, pipe := range []servicePipe{
			self.pipe.abort_read,
			self.pipe.abort_write,
		} {
			pipe.Close()
		}
	}
	return nil
}

func (self *Port) DeviceId() (major, minor uint64) {
	if self.IsOpen() {
		major, minor = GetDeviceNumber(self.stat)
	}
	return
}
func (self *Port) DeviceClassName() string {
	major, _ := self.DeviceId()
	return DeviceClassName(major)
}

func (self *Port) cancel_read() (e error) {
	if self.IsOpen() {
		e = self.pipe.abort_read.Notify()
		if e != nil { return e; }
	}
	return nil
}

func (self *Port) cancel_write() (e error) {
	if self.IsOpen() {
		e = self.pipe.abort_write.Notify()
		if e != nil { return e; }
	}
	return nil
}

func (self *Port) write(data []byte) (sent int, e error) {
	self.lock.Lock(); defer self.lock.Unlock()

	if !self.IsOpen() { return -1, PortNotOpenError; }
	data_len := len(data)
	if data_len == 0 {
		return 0, nil
	}
	var timeout time.Duration = DefaultTimeout
	var tmo float64
	var n int
	var rset, wset syscall.FdSet
	var suspect error
	t0 := time.Now()
	for sent < data_len {
		runtime.Gosched()
		
		timeout -= time.Now().Sub(t0)

		n, e = syscall.Write(int(self.fd), data[sent:])
		assert(e, "write")
		sent += n

		/*
		if timeout.is_non_blocking {
			return sent, nil // that's just fine
		}
		*/
		self.fd.FdSet(&wset)
		self.pipe.abort_write.FdSetRead(&rset)
		n = 0
		/*
		if timeout.is_infinite {
			tmo = NoSelectTimeout
			suspect = NewPortError("write failed on select")
		} else {
		*/
			if timeout.Seconds() < 0. {
				return sent, PortTimeoutError
			}
			tmo = timeout.Seconds()
			suspect = PortTimeoutError
		/*
		}
		*/
		n, e = select2(&rset, &wset, nil, tmo)
		assert(e, "select")
		if n > 0 && self.pipe.abort_write.FdIsSetRead(&rset) {
			assert(self.pipe.abort_write.Fetch(),
				"read(pipe.abort_write.r)")
			break
		}
		if n == 0 || !self.fd.FdIsSet(&wset) {
			return sent, suspect
		}
	}
	return sent, nil
}

func (self *Port) read(max int) (data []byte, e error) {
	self.lock.Lock(); defer self.lock.Unlock()

	if !self.IsOpen() { return nil, PortNotOpenError; }
	var timeout time.Duration = DefaultTimeout
	var n int
	t0 := time.Now()
	buf := make([]byte, max)
	for len(data) < max {
		runtime.Gosched()
		// TODO: ignore EAGAIN, EALREADY, EWOULDBLOCK, EINPROGRESS, EINTR
		//	 all over the code within the loop

		timeout -= time.Now().Sub(t0)

		var rset syscall.FdSet
		self.fd.FdSet(&rset)
		self.pipe.abort_read.FdSetRead(&rset)

		n, e = select2(&rset, nil, nil, timeout.Seconds())
		assert(e, "select")

		if n == 0 { // timeout
			return data, PortTimeoutError
		}

		if self.pipe.abort_read.FdIsSetRead(&rset) {
			e = self.pipe.abort_read.Fetch()
			assert(e, "read(pipe.abort_read.r)")
			break
		}

		n, e = syscall.Read(int(self.fd), buf[:max - len(data)])
		assert(e, "read")
		if n == 0 { // no data after false-positive select
			return data, NewPortError("Device disconnected or multiple access")
		}
		data = append(data, buf[:n]...)
	}
	return data, nil
}

func (self *Port) InWaiting() (n uint32, e error) {
	self.lock.Lock(); defer self.lock.Unlock()

	if !self.IsOpen() { return 0, PortNotOpenError; }
	return self.fd.TIOCINQ()
}

func (self *Port) Read(data []byte) (n int, e error) {
	self.lock.Lock(); defer self.lock.Unlock()

	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()
	// fmt.Println("Read")
	self.SetDeadline(time.Now().Add(DefaultTimeout))
	n, e = self.file.Read(data)
	if e != io.EOF {
		assert(e, "Read<%+q>(%d): %w", self.file.Name(), len(data), e)
	}
	return n, nil
}
func (self *Port) Write(data []byte) (n int, e error) {
	self.lock.Lock(); defer self.lock.Unlock()

	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()
	// fmt.Println("Write", data)
	self.SetDeadline(time.Now().Add(DefaultTimeout))
	n, e = self.file.Write(data)
	assert(e, "Write<%+q>(%d): %w", self.file.Name(), len(data), e)
	return n, nil
}
func (self *Port) SetDeadline(t time.Time) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()
	e = self.file.SetDeadline(t)
	assert(e, "SetDeadline<%+q>(%+v): %w", self.file.Name(), t, e)
	return nil
}
func (self *Port) ReadLine() (s string, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()
	var b = make([]byte, 1)
	var n int
	for b[0] != '\n' {
		runtime.Gosched()
		n, e = self.Read(b)
		assert(e, "ReadLine<%+q>: %w", self.file.Name(), e)
		if n > 0 {
			s += string(b[:1])
		}
	}
	return s, nil
}
func (self *Port) WriteLine(s string) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()
	l := len(s)
	b := make([]byte, l)
	b = []byte(s)
	var n int
	for ; l > 0; {
		runtime.Gosched()
		n, e = self.Write(b)
		assert(e, "WriteLine<%+q>(%+q): %w", self.file.Name(), string(b), e)
		l -= n
		b = b[n:]
	}
	n, e = self.file.Write([]byte("\r"))
	assert(e, "WriteLine<%+q>(%+q): %w", self.file.Name(), "\r", e)
	return nil
}
func (self *Port) hasEnd(s string, ends []string) bool {
	for _, e := range ends {
		if strings.HasSuffix(s, e) {
			return true
		}
	}
	return false
}
func (self *Port) ReadUntil(ends []string) (s string, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()
	var x string
	for {
		x, e = self.ReadLine()
		assert(e, "ReadUntil<%+q>(%+v): %w", self.file.Name(), ends, e)
		s += x
		if self.hasEnd(s, ends) {
			return s, nil
		}
	}
	return s, nil
}

type CommandChannel chan []byte

func (self *CommandChannel) SendString(s string) {
	*self <- []byte(s)
}

// Use as: `go port.Interact(commandChan, tty.Write)`
// Close the port to terminate.
func (self *Port) Interact(cc CommandChannel, write2other func([]byte) error) {
	for self.IsOpen() {
		select {
		case cmd := <- cc:
			self.Write(cmd)
		default:
			if n, e := self.InWaiting(); e != nil {
				if e != PortNotOpenError {
					assert(e, "InWaiting: %w", e)
				}
				break
			} else if n > 0 {
				b := make([]byte, n)
				if n, e := self.Read(b); e != nil {
					assert(e, "read: %w", e)
				} else if n == 0 {
					continue
				}
				if e := write2other(b); e != nil {
					assert(e, "write2other: %w", e)
				}
			}
		}
	}
}

/* EOF */
