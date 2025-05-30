package main

import (
	"errors"
	"fmt"
)

type tophError struct {
	msg string
	err error
}

func (e tophError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.err)
}

func (e tophError) Unwrap() error {
	return e.err
}

type retryableError struct {
	error
}

func (e retryableError) Unwrap() error {
	return e.error
}

type noNextPrintError struct {
	contestLocked bool
}

func (e noNextPrintError) Error() string {
	return "no next print available"
}

type printDispatchError struct {
	error
}

func (e printDispatchError) Unwrap() error {
	return e.error
}

func (e printDispatchError) Error() string {
	return fmt.Sprintf("could not dispatch print: %s", e.error.Error())
}

var (
	errInvalidToken     = errors.New("invalid token")
	errPrinterNotExist  = errors.New("printer does not exist")
	errNoDefaultPrinter = errors.New("no default printer")
)
