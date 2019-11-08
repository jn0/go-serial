package sio

import "os"
import "syscall"

// see https://github.com/pyserial/pyserial/blob/master/serial/serialposix.py

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
func (pipe *servicePipe) Fetch() (e error) {
	tmp := make([]byte, 1024)
	_, e = syscall.Read(int(pipe.rfd), tmp)
	if e != nil { return e; }
	return nil
}
func (pipe *servicePipe) Notify() (e error) {
	_, e = syscall.Write(int(pipe.wfd), []byte("x"))
	if e != nil { return e; }
	return nil
}
func (pipe *servicePipe) FdSetRead(set *syscall.FdSet) {
	pipe.rfd.FdSet(set)
}
func (pipe *servicePipe) FdSetWrite(set *syscall.FdSet) {
	pipe.wfd.FdSet(set)
}
func (pipe *servicePipe) FdClrRead(set *syscall.FdSet) {
	pipe.rfd.FdClr(set)
}
func (pipe *servicePipe) FdClrWrite(set *syscall.FdSet) {
	pipe.wfd.FdClr(set)
}
func (pipe *servicePipe) FdIsSetRead(set *syscall.FdSet) bool {
	return pipe.rfd.FdIsSet(set)
}
func (pipe *servicePipe) FdIsSetWrite(set *syscall.FdSet) bool {
	return pipe.wfd.FdIsSet(set)
}

/* EOF */
