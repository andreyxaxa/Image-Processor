package errs

import "errors"

var (
	ErrRecordNotFound   = errors.New("record not found")
	ErrUnknownOperation = errors.New("unknown operation")
)
