package cndp

import ()

type fakeCNDP struct {
	CndpInterface
}

func (c *fakeCNDP) StartSocketServer(SockAddr string) {
	return
}

func (c *fakeCNDP) CreateUdsSocket() string {
	return "/tmp/fake-socket.sock"
}

func NewFakeCndp() CndpInterface {
	return &fakeCNDP{}
}
