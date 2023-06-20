package errs

import (
	"fmt"
)

var (
	ErrCreateServer       andError_ = &andError{"Error: create server error"}
	ErrEmptyServerAddress andError_ = &andError{"Error: empty server address"}
)

type andServerError struct {
	error
	ip   string
	port string
}

func NewServerErr(msg string) error {
	return &andError{message: msg}
}

func (err *andServerError) Error() string {
	return fmt.Sprintf("handler '%s_%s': %s", err.ip, err.port, err.error.Error())
}

func (err *andServerError) Unwrap() error {
	return err.error
}

func NewCreateServerErr(ip, port string) error {
	return &andServerError{error: ErrInvlidHandlerProtocol, ip: ip, port: port}
}
