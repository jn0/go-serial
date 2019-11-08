package sio

import "syscall"
import "unsafe"
import "fmt"

const (
	TCOOFF	= 0
	TCOON	= 1
	TCIOFF	= 2
	TCION	= 3
)

func u32at(p uintptr, index uint) uint32 {
	var d *[4096]uint32 = (*[4096]uint32)(unsafe.Pointer(p))
	return d[index]
}

func (fd *Ioctl) TcFlush(qsel int) (e error) {
	e = tcFlush(*fd, qsel)
	if e != nil { return e; }
	return nil
}

func (fd *Ioctl) TcFlow(act int) (e error) {
	e = tcFlow(*fd, act)
	if e != nil { return e; }
	return nil
}

func (fd *Ioctl) SendBreak() (e error) {
	e = sendBreak(*fd)
	if e != nil { return e; }
	return nil
}

func (fd *Ioctl) TcDrain() (e error) {
	e = tcDrain(*fd)
	if e != nil { return e; }
	return nil
}

func (fd *Ioctl) InWaiting() (cnt uint32, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var data [1]uint32
	const tag = "TIOCINQ"
	const command = syscall.TIOCINQ

	// FIXME: what's returned???
	r1, r2, err := fd.ioctl(command, uintptr(unsafe.Pointer(&data)))
	assertb(r1 != UERROR, "ioctl(%v, %s, *): %v", fd, tag, err)

	cnt = u32at(r2, 0)
	return cnt, nil
}

func (fd *Ioctl) OutWaiting() (cnt uint32, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var data [1]uint32
	const tag = "TIOCOUTQ"
	const command = syscall.TIOCOUTQ

	r1, r2, err := fd.ioctl(command, uintptr(unsafe.Pointer(&data)))
	assertb(r1 != UERROR, "ioctl(%v, %s, *): %v", fd, tag, err)

	cnt = u32at(r2, 0)
	return cnt, nil
}

func (fd *Ioctl) TIOCMGET() (tiocm uint32, e error) {
	const tag = "TIOCMGET"
	const command int = syscall.TIOCMGET // FIXME: it just doesn't work

	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var data [2]uint32

	// r1, r2, err := fd.ioctl(command, uintptr(unsafe.Pointer(&data)))
	fmt.Println(	"sys=", syscall.SYS_IOCTL,
			"fd=",		uintptr(*fd),
			"call=",	uintptr(syscall.TIOCMGET),
			"data=",	uintptr(unsafe.Pointer(&data)))
	r1, r2, err := syscall.Syscall(	syscall.SYS_IOCTL,
					uintptr(*fd),
					uintptr(syscall.TIOCMGET),
					uintptr(unsafe.Pointer(&data)))
	fmt.Println(	"r1=",	r1,
			"r2=",	r2,
			"err=",	err)
	assertb(r1 != UERROR, "ioctl(%v, %s, %v): %+v", *fd, tag, &data, err)

	tiocm = u32at(r2, 0)
	return tiocm, nil
}

func (self *Port) TIOCMGET_bit(bit uint32) bool {
	if !self.IsOpen() {
		panic(PortNotOpenError)
	}

	tio, e := self.fd.TIOCMGET()
	assert(e, "TIOCMGET_bit")

	return tio & bit != 0
}

func (self *Port) CTS() bool { return self.TIOCMGET_bit(syscall.TIOCM_CTS); }
func (self *Port) DSR() bool { return self.TIOCMGET_bit(syscall.TIOCM_DSR); }
func (self *Port) RI() bool { return self.TIOCMGET_bit(syscall.TIOCM_RI); }
func (self *Port) CD() bool { return self.TIOCMGET_bit(syscall.TIOCM_CD); }

func (self *Port) SetInputFlowControl(set bool) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	if set {
		e = self.fd.TcFlow(TCION)
		assert(e, "TCION")
	} else {
		e = self.fd.TcFlow(TCIOFF)
		assert(e, "TCIOFF")
	}
	return nil
}

func (self *Port) SetOutputFlowControl(set bool) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	if set {
		e = self.fd.TcFlow(TCOON)
		assert(e, "TCOON")
	} else {
		e = self.fd.TcFlow(TCOOFF)
		assert(e, "TCOOFF")
	}
	return nil
}

func (self *Port) SendBreak() (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	e = self.fd.SendBreak()
	assert(e, "SendBreak")
	return nil
}

func (self *Port) Drain() (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	e = self.fd.TcDrain()
	assert(e, "TcDrain")
	return nil
}

func (self *Port) ResetInput() (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	e = self.fd.TcFlush(syscall.TCIFLUSH)
	assert(e, "TcFlush.TCIFLUSH")
	return nil
}

func (self *Port) ResetOutput() (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	e = self.fd.TcFlush(syscall.TCOFLUSH)
	assert(e, "TcFlush.TCOFLUSH")
	return nil
}
/* EOF */
