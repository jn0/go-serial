package sio

import (
	"fmt"
	"strings"
	"github.com/go-errors/errors"
)

type PortError struct {
	message string
	wrapped error
}
func (self *PortError) Error() string {
	if self.wrapped != nil {
		return fmt.Sprintf("%s: %v", self.message, self.wrapped)
	}
	return self.message
}
func (self *PortError) Unwrap() error {
	return self.wrapped
}
func (self *PortError) Set(e error, m string) {
	self.message = m
	self.wrapped = e
}

var PortNotOpenError = NewPortError("Port was not open")
var PortTimeoutError = NewPortError("Port timed out")

func NewPortError(message string, args ...interface{}) *PortError {
	var pe *PortError = &PortError{}
	if strings.Contains(message, "%w") {
		for _, x := range args {
			switch x.(type) {
			case error, errors.Error:
				ge := errors.Errorf(message, args...)
				pe.Set(x.(error), ge.Error())
				return pe
			}
		}
		panic(errors.Errorf("No error object in %#v", args))
	} else {
		pe.Set(nil, fmt.Sprintf(message, args...))
		return pe
	}
}

func assert(e error, args ...interface{}) {
	if e != nil {
		if args == nil || len(args) == 0 {
			panic(NewPortError("Port error: %w", e))
		}

		var message string
		message, args = args[0].(string), args[1:]
		if args == nil || len(args) == 0 {
			panic(NewPortError("%s: %w", message, e))
		}
		panic(NewPortError(message, args...))
	}
}
func assertb(ok bool, args ...interface{}) {
	if !ok {
		e := NewPortError("Assertion failed")
		if args != nil && len(args) > 0 {
			var message string
			message, args = args[0].(string), args[1:]
			if args == nil || len(args) == 0 {
				e = NewPortError("Assertion failed: %s", message)
			} else {
				e = NewPortError("Assertion failed: " + message, args...)
			}
		}
		panic(e)
	}
}

func WrapError(err interface{}) *errors.Error { // github.com/dsoprea/go-logging
    es, ok := err.(*errors.Error)
    if ok {
        return es
    }
    return errors.Wrap(err, 1)
}


/* EOF */
