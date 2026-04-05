package api

import "errors"

var (
	ErrConnectionFailed = errors.New("connection failed")
	ErrInvalidQuery     = errors.New("invalid query")
	ErrNotSupported     = errors.New("not supported")
	ErrBlockedQuery     = errors.New("query blocked by security engine")
)
