package errs

import (
	"fmt"
)

var (
	ErrInvlidHandlerProtocol andError_ = &andError{"Error: invlalid protocol name"}
)

type andHandlerError struct {
	error
	proto string
	port  string
}

func NewHandlerErr(msg string) error {
	return &andError{message: msg}
}

func (err *andHandlerError) Error() string {
	return fmt.Sprintf("handler '%s_%s': %s", err.proto, err.port, err.error.Error())
}

func (err *andHandlerError) Unwrap() error {
	return err.error
}

func NewInvlidHandlerProtocolErr(proto, port string) error {
	return &andHandlerError{error: ErrInvlidHandlerProtocol, proto: proto, port: port}
}
