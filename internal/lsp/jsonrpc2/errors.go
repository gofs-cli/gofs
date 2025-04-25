package jsonrpc2

import "errors"

var (
	ErrInvalidParams = errors.New("invalid params")
	ErrInternalError = errors.New("internal error")
)
