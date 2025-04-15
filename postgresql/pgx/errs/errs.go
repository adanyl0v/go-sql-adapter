package errs

type Error struct {
	msg string
	err error
}

func New(message string, err error) Error {
	return Error{
		msg: message,
		err: err,
	}
}

func (e Error) Error() string {
	return e.msg
}

func (e Error) Unwrap() error {
	return e.err
}
