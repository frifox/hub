package encoder

import "context"

type Devices struct {
	Cam Device
	PC  Device
}

type Device struct {
	Address string
	//IP string
	//Port string

	Name string
	Channel string

	context.Context
	context.CancelFunc
}

func (d *Device) Run() {
	// TODO
}