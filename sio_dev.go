package sio

import (
	"os"
	"fmt"
	"bufio"
	"sync"
	"strconv"
	"strings"
)

const ProcDevices = "/proc/devices"
/*
Character devices:
  1 mem
  4 tty
  5 /dev/tty
  5 /dev/console
  ...
251 dimmctl
252 ndctl
253 tpm
254 gpiochip

Block devices:
  7 loop
  8 sd
  9 md
 11 sr
 65 sd
 ...
253 device-mapper
254 mdp
259 blkext
*/
const ProcMisc = "/proc/misc"
/*
234 btrfs-control
232 kvm
235 autofs
 56 memory_bandwidth
 57 network_throughput
 58 network_latency
 59 cpu_dma_latency
184 microcode
227 mcelog
236 device-mapper
223 uinput
  1 psaux
200 tun
237 loop-control
 60 lightnvm
183 hw_random
228 hpet
229 fuse
 61 ecryptfs
231 snapshot
 62 rfkill
 63 vga_arbiter
*/
var AutoLoad = []string{
	ProcDevices,
	ProcMisc,
}

type DeviceClassMapper map[int][]string

func (self *DeviceClassMapper) add(value int, name string) {
	(*self)[value] = append((*self)[value], strings.TrimSpace(name))
}
func (self *DeviceClassMapper) load(file string) (e error) {
	if !exists(file) {
		return os.ErrNotExist
	}
	f, e := os.Open(file)
	if e != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var value, ln int64
	for scanner.Scan() {
		line := scanner.Text()
		ln += 1
		if len(line) <= 4 || line[3] != ' ' {
			continue // comments, etc
		}
		value, e = strconv.ParseInt(strings.TrimSpace(line[:3]), 10, 16)
		if e != nil {
			fmt.Fprintf(os.Stderr, "%s:%d: %v", file, ln, e)
			continue // shit happens
		}
		self.add(int(value), line[4:])
	}
	e = scanner.Err()

	return nil
}
func (self *DeviceClassMapper) GetName(class int) (name string) {
	var found bool
	var names []string
	names, found = (*self)[class]
	if found {
		name = strings.Join(names, ";")
	} else {
		name = fmt.Sprintf("<deviceClass#%v>", class)
	}
	return
}
func (self *DeviceClassMapper) Load() {
	if x, ok := (*self)[-31415]; ok && strings.Join(x, "/") == "JNO" {
		return // already loaded
	}
	*self = make(map[int][]string)
	self.add(-31415, "JNO")

	for _, name := range AutoLoad {
		self.load(name)
	}
}

var lock sync.Mutex
func NewDeviceClassMapper() *DeviceClassMapper {
	lock.Lock() ; defer lock.Unlock()
	var dcm *DeviceClassMapper = &DeviceClassMapper{}
	dcm.Load()
	return dcm
}

var deviceClassMapper = NewDeviceClassMapper() // still may have race condition

func DeviceClassName(class int) string {
	return deviceClassMapper.GetName(class)
}

/* EOF */
