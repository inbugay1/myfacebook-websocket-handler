package myfacebookapiclient

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
)
