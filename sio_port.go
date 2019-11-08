package sio

import (
	"os"
	"io"
	"time"
	"strings"
	"syscall"
)
// import "fmt"

const USB0 = "/dev/ttyUSB0"
const PortOpenFlags = os.O_RDWR | syscall.O_NOCTTY | syscall.O_NONBLOCK
const DefaultTimeout = time.Millisecond * 50 // 0.05 sec

func exists(path string) bool {
	_, e := os.Stat(path)
	return e == nil || !os.IsNotExist(e)
}

type servicePipe struct {
	r, w *os.File
	rfd, wfd Ioctl
}
func (pipe *servicePipe) Open() (e error) {
	pipe.r, pipe.w, e = os.Pipe()
	if e != nil { return e; }
	pipe.rfd.Set(pipe.r.Fd())
	pipe.wfd.Set(pipe.w.Fd())
	e = pipe.rfd.NonBlock(true)
	if e != nil { return e; }
	return nil
}
func (pipe *servicePipe) Close() {
	pipe.r.Close()
	pipe.r = nil
	pipe.rfd = ZeroIoctl
	pipe.w.Close()
	pipe.w = nil
	pipe.wfd = ZeroIoctl
}
func (pipe *servicePipe) Read() (e error) {
	tmp := make([]byte, 1024)
	_, e = syscall.Read(int(pipe.rfd), tmp)
	if e != nil { return e; }
	return nil
}

type Port struct {
	file *os.File
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

	assertb(exists(path), "Port.Open(%+q): no path", path)
	assertb(!self.IsOpen(), "Port.Open(%+q): *object* is already open", path)

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

	e = self.fd.Reconfigure(&self.termios, self)
	assert(e, "Open: cannot configure port")

	if !self.dsrdtr {
		self.fd.SetDTR(false)
	}
	if !self.rtscts {
		self.fd.SetRTS(false)
	}
	e = self.ResetInput()
	assert(e, "ResetInput")

	e = self.pipe.abort_read.Open()
	assert(e, "pipe(read)")

	e = self.pipe.abort_write.Open()
	assert(e, "pipe(write)")

	return nil
}
func (self *Port) Close() error {
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

func fd_set(fd_orig uintptr, set *syscall.FdSet) {
	const bits = 64
	var fd = uint64(fd_orig)
	var clause = fd / bits
	if clause > 15 { panic("fd too big"); }
	var bit int64 = 1 << (fd % bits)
	set.Bits[clause] |= bit
}

func fd_isset(fd_orig uintptr, set *syscall.FdSet) bool {
	const bits = 64
	var fd = uint64(fd_orig)
	var clause = fd / bits
	if clause > 15 { panic("fd too big"); }
	var bit int64 = 1 << (fd % bits)
	return set.Bits[clause] & bit == bit
}

func fd_count(set *syscall.FdSet) (n int) {
	var i uintptr
	for i = 0; i < 64 * 16; i++ {
		if fd_isset(i, set) { n++; }
	}
	return n
}

func timeval(seconds float64) (res syscall.Timeval) {
	res.Sec = int64(seconds)
	res.Usec = int64(seconds * 1000000.) % 1000000
	return res
}

func (self *Port) cancel_read() (e error) {
	if self.IsOpen() {
		_, e = self.pipe.abort_read.w.Write([]byte("x"))
		if e != nil { return e; }
	}
	return nil
}

func (self *Port) cancel_write() (e error) {
	if self.IsOpen() {
		_, e = self.pipe.abort_write.w.Write([]byte("x"))
		if e != nil { return e; }
	}
	return nil
}

func (self *Port) write(data []byte) (sent int, e error) {
	if !self.IsOpen() { return -1, PortNotOpenError; }
	data_len := len(data)
	if data_len == 0 {
		return 0, nil
	}
	var timeout time.Duration = DefaultTimeout
	var n int
	var rset, wset syscall.FdSet
	var suspect error
	var tmo syscall.Timeval
	var tmop *syscall.Timeval
	t0 := time.Now()
	for sent < data_len {
		
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
		self.pipe.abort_write.rfd.FdSet(&rset)
		n = 0
		/*
		if timeout.is_infinite {
			tmop = nil
			suspect = NewPortError("write failed on select")
		} else {
		*/
			if timeout.Seconds() < 0. {
				return sent, PortTimeoutError
			}
			tmo = timeval(timeout.Seconds())
			tmop = &tmo
			suspect = PortTimeoutError
		/*
		}
		*/
		n, e = syscall.Select(fd_count(&rset) + fd_count(&wset),
					&rset, &wset, nil, tmop)
		assert(e, "select")
		if n > 0 && self.pipe.abort_write.rfd.FdIsSet(&rset) {
			e = self.pipe.abort_write.Read()
			assert(e, "read(pipe.abort_write.r)")
			break
		}
		if n == 0 || !self.fd.FdIsSet(&wset) {
			return sent, suspect
		}
	}
	return sent, nil
}

func (self *Port) read(max int) (data []byte, e error) {
	if !self.IsOpen() { return nil, PortNotOpenError; }
	var timeout time.Duration = DefaultTimeout
	var n int
	t0 := time.Now()
	buf := make([]byte, max)
	for len(data) < max {
		// TODO: ignore EAGAIN, EALREADY, EWOULDBLOCK, EINPROGRESS, EINTR
		//	 all over the code within the loop

		timeout -= time.Now().Sub(t0)
		tmo := timeval(timeout.Seconds())

		var rset syscall.FdSet
		fd_set(uintptr(self.fd), &rset)
		fd_set(uintptr(self.pipe.abort_read.rfd), &rset)
		n, e = syscall.Select(fd_count(&rset), &rset, nil, nil, &tmo)
		assert(e, "select")
		if n == 0 { // timeout
			return data, PortTimeoutError
		}

		if fd_isset(uintptr(self.pipe.abort_read.rfd), &rset) {
			e = self.pipe.abort_read.Read()
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
func (self *Port) Read(data []byte) (n int, e error) {
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
	for ; b[0] != '\n'; {
		// time.Sleep(10 * time.Millisecond)
		n, e = self.Read(b)
		// fmt.Println("Read:", n, e)
		assert(e, "ReadLine<%+q>: %w", self.file.Name(), e)
		if n > 0 {
			s += string(b[:1])
			// fmt.Printf("n=%v s=%+q\n", n, s)
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

/* EOF */