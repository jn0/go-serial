package sio

import "syscall"

const CMSPAR uint32 = 010000000000 // linux, octal!
const CRTSCTS uint32 = 020000000000	/* flow control */

type CharSize uint
const (
	CHAR_SIZE_5 = CharSize(5)
	CHAR_SIZE_6 = CharSize(6)
	CHAR_SIZE_7 = CharSize(7)
	CHAR_SIZE_8 = CharSize(8)
)
func IsValidCharSize(v int) bool {
	switch CharSize(uint(v)) {
	case CHAR_SIZE_5: return true
	case CHAR_SIZE_6: return true
	case CHAR_SIZE_7: return true
	case CHAR_SIZE_8: return true
	}
	return false
}
var SysCharSize = map[CharSize]uint32{
	CHAR_SIZE_5: syscall.CS5,
	CHAR_SIZE_6: syscall.CS6,
	CHAR_SIZE_7: syscall.CS7,
	CHAR_SIZE_8: syscall.CS8,
}

type Parity uint
const (
	PARITY_NONE = Parity(0)
	PARITY_ODD = Parity(1)
	PARITY_EVEN = Parity(2)
	PARITY_MARK = Parity(3)
	PARITY_SPACE = Parity(4)
)
func IsValidParity(v int) bool {
	switch Parity(uint(v)) {
	case PARITY_NONE: return true
	case PARITY_ODD: return true
	case PARITY_EVEN: return true
	case PARITY_MARK: return true
	case PARITY_SPACE: return true
	}
	return false
}

type StopBits uint
const (
	STOP_BITS_1 = StopBits(1)
	STOP_BITS_2 = StopBits(2)
)
func IsValidStopBits(v int) bool {
	switch StopBits(uint(v)) {
	case STOP_BITS_1: return true
	case STOP_BITS_2: return true
	}
	return false
}

type BitRate uint32
const (
	BIT_RATE_B0 = BitRate(syscall.B0)
	BIT_RATE_B50 = BitRate(syscall.B50)
	BIT_RATE_B75 = BitRate(syscall.B75)
	BIT_RATE_B110 = BitRate(syscall.B110)
	BIT_RATE_B134 = BitRate(syscall.B134)
	BIT_RATE_B150 = BitRate(syscall.B150)
	BIT_RATE_B200 = BitRate(syscall.B200)
	BIT_RATE_B300 = BitRate(syscall.B300)
	BIT_RATE_B600 = BitRate(syscall.B600)
	BIT_RATE_B1200 = BitRate(syscall.B1200)
	BIT_RATE_B1800 = BitRate(syscall.B1800)
	BIT_RATE_B2400 = BitRate(syscall.B2400)
	BIT_RATE_B4800 = BitRate(syscall.B4800)
	BIT_RATE_B9600 = BitRate(syscall.B9600)
	BIT_RATE_B19200 = BitRate(syscall.B19200)
	BIT_RATE_B38400 = BitRate(syscall.B38400)
	BIT_RATE_B57600 = BitRate(syscall.B57600)
	BIT_RATE_B115200 = BitRate(syscall.B115200)
	BIT_RATE_B230400 = BitRate(syscall.B230400)
	BIT_RATE_B460800 = BitRate(syscall.B460800)
	BIT_RATE_B500000 = BitRate(syscall.B500000)
	BIT_RATE_B576000 = BitRate(syscall.B576000)
	BIT_RATE_B921600 = BitRate(syscall.B921600)
	BIT_RATE_B1000000 = BitRate(syscall.B1000000)
	BIT_RATE_B1152000 = BitRate(syscall.B1152000)
	BIT_RATE_B1500000 = BitRate(syscall.B1500000)
	BIT_RATE_B2000000 = BitRate(syscall.B2000000)
	BIT_RATE_B2500000 = BitRate(syscall.B2500000)
	BIT_RATE_B3000000 = BitRate(syscall.B3000000)
	BIT_RATE_B3500000 = BitRate(syscall.B3500000)
	BIT_RATE_B4000000 = BitRate(syscall.B4000000)
)
func IsValidSpeed(v uint32) bool {
	switch BitRate(v) {
	case BIT_RATE_B0: return true
	case BIT_RATE_B50: return true
	case BIT_RATE_B75: return true
	case BIT_RATE_B110: return true
	case BIT_RATE_B134: return true
	case BIT_RATE_B150: return true
	case BIT_RATE_B200: return true
	case BIT_RATE_B300: return true
	case BIT_RATE_B600: return true
	case BIT_RATE_B1200: return true
	case BIT_RATE_B1800: return true
	case BIT_RATE_B2400: return true
	case BIT_RATE_B4800: return true
	case BIT_RATE_B9600: return true
	case BIT_RATE_B19200: return true
	case BIT_RATE_B38400: return true
	case BIT_RATE_B57600: return true
	case BIT_RATE_B115200: return true
	case BIT_RATE_B230400: return true
	case BIT_RATE_B460800: return true
	case BIT_RATE_B500000: return true
	case BIT_RATE_B576000: return true
	case BIT_RATE_B921600: return true
	case BIT_RATE_B1000000: return true
	case BIT_RATE_B1152000: return true
	case BIT_RATE_B1500000: return true
	case BIT_RATE_B2000000: return true
	case BIT_RATE_B2500000: return true
	case BIT_RATE_B3000000: return true
	case BIT_RATE_B3500000: return true
	case BIT_RATE_B4000000: return true
	}
	return false
}

func setTermios(termios *syscall.Termios, port *Port) (e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var vmin, vtime uint8
	if port.inter_byte_timeout > 0. {
		vmin = 1
		vtime = uint8(port.inter_byte_timeout * 10.)
	}

	termios.Lflag &= ^uint32(syscall.ICANON | syscall.ECHO    | syscall.ECHOE |
				 syscall.ECHOK  | syscall.ECHONL  | syscall.ISIG  |
				 syscall.IEXTEN | syscall.ECHOCTL | syscall.ECHOKE)

	termios.Oflag &= ^uint32(syscall.OPOST | syscall.ONLCR | syscall.OCRNL)

	termios.Iflag &= ^uint32(syscall.INLCR  | syscall.IGNCR | syscall.ICRNL |
				 syscall.IGNBRK | syscall.IUCLC | syscall.PARMRK|
				 syscall.INPCK  | syscall.ISTRIP)
	switch port.parity {
	case PARITY_NONE:
		termios.Cflag &= ^uint32(syscall.PARENB | syscall.PARODD | CMSPAR)
	case PARITY_ODD:
		termios.Cflag &= ^uint32(CMSPAR)
		termios.Cflag |= syscall.PARENB | syscall.PARODD
	case PARITY_EVEN:
		termios.Cflag &= ^uint32(syscall.PARODD | CMSPAR)
		termios.Cflag |= syscall.PARENB
	case PARITY_MARK:
		if CMSPAR != 0 {
			termios.Cflag |= syscall.PARENB | syscall.PARODD | CMSPAR
		}
	case PARITY_SPACE:
		if CMSPAR != 0 {
			termios.Cflag |= syscall.PARENB | CMSPAR
			termios.Cflag &= ^uint32(syscall.PARODD)
		}
	}

	if port.xonxoff {
		termios.Iflag |= syscall.IXON | syscall.IXOFF
	} else {
		termios.Iflag &= ^uint32(syscall.IXON | syscall.IXOFF | syscall.IXANY)
	}
	if port.rtscts {
		termios.Cflag |= CRTSCTS
	} else {
		termios.Cflag &= ^uint32(CRTSCTS)
	}

	termios.Ispeed = uint32(port.speed)
	termios.Ospeed = uint32(port.speed)

	termios.Cflag |= syscall.CLOCAL | syscall.CREAD
	termios.Cflag &= ^uint32(syscall.CSIZE)
	termios.Cflag |= SysCharSize[port.char_size]
	switch port.stop_bits {
	case STOP_BITS_1: termios.Cflag &= ^uint32(syscall.CSTOPB)
	case STOP_BITS_2: termios.Cflag |= syscall.CSTOPB
	}

	termios.Cc[syscall.VMIN] = vmin
	termios.Cc[syscall.VTIME] = vtime
	return nil
}

/* EOF */
