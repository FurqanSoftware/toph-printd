package main

import "github.com/rs/xid"

func newID() string {
	return xid.New().String()
}
