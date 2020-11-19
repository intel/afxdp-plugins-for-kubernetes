package cndp

import ()

type fakeCNDP struct {
	Interface
}

func (c *fakeCNDP) StartSocketServer(SockAddr string) {
	return
}

func (c *fakeCNDP) CreateUdsSocket() string {
	return "/tmp/fake-socket.sock"
}

/*
NewFakeCndp returns a fake CNDP object of type cndp.Interface.
The functions of the fake CNDP have little functionality. Returns are static and predictable.
This is used while testing in other areas of the device plugin.
*/
func NewFakeCndp() Interface {
	return &fakeCNDP{}
}
