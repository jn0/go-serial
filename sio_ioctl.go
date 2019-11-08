package sio

import (
	"syscall"
	"unsafe"
)
//import "fmt"

const UERROR = uintptr(18446744073709551615)

type Ioctl uintptr

const ZeroIoctl = Ioctl(0)
const ASYNC_LOW_LATENCY uint32 = 0x2000
const E_OK = syscall.Errno(0)

func (fd *Ioctl) Set(value uintptr) {
	*fd = Ioctl(value)
}

func (fd *Ioctl) ioctl(a1 int, a2 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall(syscall.SYS_IOCTL, uintptr(*fd), uintptr(a1), a2)
}
func (fd *Ioctl) fcntl(a1 int, a2 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	return syscall.Syscall(syscall.SYS_FCNTL, uintptr(*fd), uintptr(a1), a2)
}

func (fd *Ioctl) FdSet(set *syscall.FdSet) {
	fd_set(uintptr(*fd), set)
}
func (fd *Ioctl) FdClr(set *syscall.FdSet) {
	fd_clr(uintptr(*fd), set)
}
func (fd *Ioctl) FdIsSet(set *syscall.FdSet) bool {
	return fd_isset(uintptr(*fd), set)
}

func (fd *Ioctl) Flock(how int) (e error) {
	return syscall.Flock(int(*fd), how)
}

func (fd *Ioctl) GETFL() (flags uintptr, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var buf [1]uint32
	flags, _, err := fd.fcntl(syscall.F_GETFL, uintptr(unsafe.Pointer(&buf)))
	assertb(err == E_OK, "F_GETFL")
	return flags, nil
}

func (fd *Ioctl) NonBlock(set bool) (e error) {
	fl, e := fd.GETFL()
	if set {
		fl |= syscall.O_NONBLOCK
	} else {
		fl &= ^uintptr(syscall.O_NONBLOCK)
	}

	var buf [1]uint32
	buf[0] = uint32(fl)
	_, _, err := fd.fcntl(syscall.F_SETFL, uintptr(unsafe.Pointer(&buf)))
	assertb(err == E_OK, "F_SETFL")
	return nil
}

func (fd *Ioctl) get_low_latency_mode() (mode bool, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var buf [32]uint32
	_, _, err := fd.ioctl(syscall.TIOCGSERIAL, uintptr(unsafe.Pointer(&buf)))
	if err != E_OK {
		return false, NewPortError("ioctl(%v, TIOCGSERIAL, *): %v", fd, err)
	}
	return buf[4] & ASYNC_LOW_LATENCY == ASYNC_LOW_LATENCY, nil
}
func (fd *Ioctl) set_low_latency_mode(set bool) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var buf [32]uint32
	_, _, err := fd.ioctl(syscall.TIOCGSERIAL, uintptr(unsafe.Pointer(&buf)))
	assertb(err == E_OK, "ioctl(%v, TIOCGSERIAL, *): %v", fd, err)
	if set {
		buf[4] |= ASYNC_LOW_LATENCY
	} else {
		buf[4] &= ^ASYNC_LOW_LATENCY
	}
	_, _, err = fd.ioctl(syscall.TIOCSSERIAL, uintptr(unsafe.Pointer(&buf)))
	assertb(err == E_OK, "ioctl(%v, TIOCSSERIAL, *): %v", fd, err)
	return nil
}

func (fd *Ioctl) TcGetAttr() (termios syscall.Termios, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	_, _, err := fd.ioctl(syscall.TCGETS, uintptr(unsafe.Pointer(&termios)))
	assertb(err == E_OK, "ioctl(%v, TCGETS, *): %v", fd, err)
	return termios, nil
}
func (fd *Ioctl) TcSetAttr(termios syscall.Termios) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	_, _, err := fd.ioctl(syscall.TCSETS, uintptr(unsafe.Pointer(&termios)))
	assertb(err == E_OK, "ioctl(%v, TCSETS, *): %v", fd, err)
	return nil
}

func (fd *Ioctl) TIOCGRS485(buf *[8]uint32) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	_, _, err := fd.ioctl(syscall.TIOCGRS485, uintptr(unsafe.Pointer(buf)))
	assertb(err == E_OK, "ioctl(%v, TIOCGRS485, *): %v", fd, err)
	return nil
}
func (fd *Ioctl) TIOCSRS485(buf [8]uint32) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	_, _, err := fd.ioctl(syscall.TIOCSRS485, uintptr(unsafe.Pointer(&buf)))
	assertb(err == E_OK, "ioctl(%v, TIOCSRS485, *): %v", fd, err)
	return nil
}

func (fd *Ioctl) SetRs485(rs485 Rs485) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var buf [8]uint32

	e = fd.TIOCGRS485(&buf)
	assert(e, "TIOCGRS485")
	rs485.Update(&buf)
	e = fd.TIOCSRS485(buf)
	assert(e, "TIOCSRS485")
	return nil
}

func (fd *Ioctl) Reconfigure(termios *syscall.Termios, port *Port) (e error) {
	// https://github.com/pyserial/pyserial/blob/master/serial/serialposix.py
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	if port.exclusive {
		e = fd.Flock(syscall.LOCK_EX | syscall.LOCK_NB)
		assert(e, "Flock(LOCK_EX|LOCK_NB)")
		e = fd.Flock(syscall.LOCK_UN)
		assert(e, "Flock(LOCK_UN)")
	}

	*termios, e = fd.TcGetAttr()
	assert(e, "TCGETATTR")

	var trms syscall.Termios = *termios

	e = setTermios(&trms, port)
	assert(e, "setTermios")

	e = fd.TcSetAttr(trms)
	assert(e, "TCSETATTR")

	if port.rs485.Enabled {
		e = fd.SetRs485(port.rs485)
		assert(e, "fd.set_rs485(%+v): %w", port.rs485, e)
	}

	return nil
}

func (fd *Ioctl) SetRTS(set bool) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var data [1]uint32
	var tag string
	var command int

	if set {
		tag = "TIOCMBIS"
		command = syscall.TIOCMBIS
	} else {
		tag = "TIOCMBIC"
		command = syscall.TIOCMBIC
	}

	data[0] = syscall.TIOCM_RTS
	_, _, err := fd.ioctl(command, uintptr(unsafe.Pointer(&data)))
	assertb(err == E_OK, "ioctl(%v, %s, *): %v", fd, tag, err)
	return nil
}
func (fd *Ioctl) SetDTR(set bool) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var data [1]uint32
	var tag string
	var command int

	if set {
		tag = "TIOCMBIS"
		command = syscall.TIOCMBIS
	} else {
		tag = "TIOCMBIC"
		command = syscall.TIOCMBIC
	}

	data[0] = syscall.TIOCM_DTR
	_, _, err := fd.ioctl(command, uintptr(unsafe.Pointer(&data)))
	assertb(err == E_OK, "ioctl(%v, %s, *): %v", fd, tag, err)
	return nil
}

/* EOF */
