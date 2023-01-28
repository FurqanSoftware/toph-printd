package main

import "fmt"

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
