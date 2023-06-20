package errs

var (
	ErrEmptyHandlerProtocol andError_ = &andError{"Error: empty protocol name"}
)

type andError_ interface {
	error
}

type andError struct {
	message string
}

func New(msg string) error {
	return &andError{message: msg}
}

func (err *andError) Error() string {
	return err.message
}

func (err *andError) Unwrap() error {
	return err
}
