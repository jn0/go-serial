// Copyright (C) jno, 2019
// Support for select(2) via syscall.Select()
// See also: Ioctl.Fd*() methods
package sio

import "syscall"

const NoSelectTimeout float64 = -1. // will cause timeout pointer to be nil

var ZeroFdSet = syscall.FdSet{
	Bits: [16]int64{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0},
}

func _fd_bit(fd_orig uintptr) (clause uint64, bit int64) {
	const bits = 64
	var fd = uint64(fd_orig)
	clause = fd / bits
	if clause > 15 { panic("fd too big"); }
	bit = 1 << (fd % bits)
	return clause, bit
}

func fd_zero(set *syscall.FdSet) {
	*set = ZeroFdSet
	/*
	for i := 0; i < len(set.Bits); i++ {
		set.Bits[i] = 0
	}
	*/
}

func fd_clr(fd_orig uintptr, set *syscall.FdSet) {
	clause, bit := _fd_bit(fd_orig)
	set.Bits[clause] &= ^bit
}

func fd_set(fd_orig uintptr, set *syscall.FdSet) {
	clause, bit := _fd_bit(fd_orig)
	set.Bits[clause] |= bit
}

func fd_isset(fd_orig uintptr, set *syscall.FdSet) bool {
	clause, bit := _fd_bit(fd_orig)
	return set.Bits[clause] & bit == bit
}

func fd_count(set *syscall.FdSet) (n int) {
	var i uintptr
	if set == nil {
		return 0
	}
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

// The wrapper
func select2(r, w, x *syscall.FdSet, seconds float64) (n int, e error) {
	defer func() {
		if state := recover(); state != nil {
			e = WrapError(state.(error))
		}
	}()

	var tmo syscall.Timeval
	var tmo_ptr *syscall.Timeval
	if seconds == NoSelectTimeout {
		tmo_ptr = nil
	} else {
		tmo = timeval(seconds)
		tmo_ptr = &tmo
	}
	n, e = syscall.Select(fd_count(r) + fd_count(w) + fd_count(x),
				r, w, x, tmo_ptr)
	assert(e, "select")
	return n, nil
}

/* EOF */
