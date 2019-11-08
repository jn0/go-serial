package sio
// https://github.com/npat-efault/serial/blob/master/termios/cgo_termios.go

/*
#include <termios.h>
#include <unistd.h>
*/
import "C"
import "syscall"

// Flow suspends or resumes the transmission or reception of data on
// the terminal associated with fd, depending on the value of the act
// argument. The act argument value must be one of: TCOOFF (suspend
// transmission), TCOON (resume transmission), TCIOFF (suspend
// reception by sending a STOP char), TCION (resume reception by
// sending a START char). See also tcflow(3).
func tcFlow(fd Ioctl, act int) error {
	for {
		r, err := C.tcflow(C.int(fd), C.int(act))
		if r < 0 {
			// This is most-likely not possible, but
			// better be safe.
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		return nil
	}
}

// SendBreak sends a continuous stream of zero bits to the terminal
// corresponding to file-descriptor fd, lasting between 0.25 and 0.5
// seconds.
func sendBreak(fd Ioctl) error {
	r, err := C.tcsendbreak(C.int(fd), 0)
	if r < 0 {
		return err
	}
	return nil
}

// Drain blocks until all data written to the terminal fd are
// transmitted. See also tcdrain(3). If the system call is interrupted
// by a signal, Drain retries it automatically.
func tcDrain(fd Ioctl) error {
	for {
		r, err := C.tcdrain(C.int(fd))
		if r < 0 {
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		return nil
	}
}

// Flush discards data received but not yet read (input queue), and/or
// data written but not yet transmitted (output queue), depending on
// the value of the qsel argument. Argument qsel must be one of the
// constants TCIFLUSH (flush input queue), TCOFLUSH (flush output
// queue), TCIOFLUSH (flush both queues). See also tcflush(3).
func tcFlush(fd Ioctl, qsel int) error {
	for {
		r, err := C.tcflush(C.int(fd), C.int(qsel))
		if r < 0 {
			// This is most-likely not possible, but
			// better be safe.
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		return nil
	}
}

/* EOF */
