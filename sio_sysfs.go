package sio

import (
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
	"strconv"
)
import "fmt"

const SysfsRoot = "/sys"
const SysfsClass = "tty"

type SysFS uint64

func (self *SysFS) Use(rdev uint64) {
	*self = SysFS(rdev)
}
func (self *SysFS) locate(startpath string, major, minor uint64, seen map[string]bool) []string {
	var isSymLink = func(info os.FileInfo) bool {
		return info.Mode() & os.ModeSymlink == os.ModeSymlink
	}

	var ls, symlinks []string

	s, k := seen[startpath]
	if k && s {
		return ls
	}
	seen[startpath] = true
	// fmt.Println("walking", startpath)
	_ = filepath.Walk(startpath, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err; }

		s, k := seen[path]
		if k && s {
			return nil
		}
		seen[path] = true

		if info.IsDir() {
			// fmt.Println("DIR", path)
			return nil
		}
		if isSymLink(info) {
			rpath, e := readlink(path)
			st, e := os.Stat(rpath)
			if e != nil {
				fmt.Println("Oops(stat", rpath, "):", e)
				return nil
			}
			if st.IsDir() {
				symlinks = append(symlinks, rpath)
			}
		} else if filepath.Base(path) == "dev" {
			bytes, e := ioutil.ReadFile(path)
			if e != nil {
				fmt.Println("Read(", path, "):", e)
				return nil
			}
			text := string(bytes)
			tmp := strings.Split(strings.TrimSpace(text), ":")
			assertb(len(tmp) == 2, "Bad %+q: %+q", path, text)

			mjr, e := strconv.ParseInt(tmp[0], 10, 16)
			if e != nil {
				fmt.Fprintf(os.Stderr, "%s: %v", path, text)
				return nil

			}
			mnr, e := strconv.ParseInt(tmp[1], 10, 16)
			if e != nil {
				fmt.Fprintf(os.Stderr, "%s: %v", path, text)
				return nil
			}
			if major == uint64(mjr) && minor == uint64(mnr) {
				// fmt.Println("has dev", path)
				ls = append(ls, filepath.Dir(path))
			}
		}
		return nil
	})
	for _, p := range symlinks {
		tmp := self.locate(p, major, minor, seen)
		ls = append(ls, tmp...)
	}
	return ls
}
func (self *SysFS) Locate(class string, major, minor uint64) []string {
	return self.locate(filepath.Join(SysfsRoot, "class", class),
			   major, minor,
			   make(map[string]bool))
}

func readlink(link string) (path string, e error) {
	rpath, e := os.Readlink(link)
	if e != nil {
		fmt.Println("Oops! At readlink", path, ":", e)
		return
	}
	if filepath.IsAbs(rpath) {
		path = rpath
	} else {
		path = filepath.Join(filepath.Dir(link), rpath)
	}
	return
}

func PrintLocations(locations []string,
		    printf func(string, ...interface{}) (int, error),
) {
	var n int
	for i, v := range locations {
		printf("%3d:\t%#v\n", i + 1, v)
		n++
	}
	if n == 0 {
		printf("%3d:\tnone\n", 0)
	}
}

/* EOF */
