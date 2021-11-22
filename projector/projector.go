package projector

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

//type Devices map[string]Device

type Device struct {
	Address string
	Name string

	context.Context
	context.CancelFunc

	commands chan Command
	State State

	EventCallbacks map[string]chan interface{}
}

type Callback func(string, interface{})

type State struct {
	Power
	Freeze
	Blank
}

func (d *Device) Register(clientID string, callback chan interface{}) {
	fmt.Printf("[Projector:%s] Registering callbacks. Client:%s\n", d.Address, clientID)

	if d.EventCallbacks == nil {
		d.EventCallbacks = make(map[string]chan interface{})
	}

	d.EventCallbacks[clientID] = callback
}
func (d *Device) Deregister(id string) {
	delete(d.EventCallbacks, id)
}

func (d *Device) Run() {
	fmt.Printf("Running Projector(%s)\n", d.Address)
	for {
		err := d.Connect()
		if err != nil {
			fmt.Printf("ERROR projector.Connect(%s): %v\n", d.Address, err)
			time.Sleep(time.Second * 5)
			continue
		}

		//go d.reader()
		//go d.writer()

		<- d.Done()
	}
}
func (d *Device) Connect() (err error) {
	//fmt.Printf("projector.Connect(%s)\n", d.Address)

	conn, err := net.DialTimeout("tcp", d.Address, time.Second * 1)
	if err != nil {
		return fmt.Errorf("DialTCP: %v", err)
	} else {
		defer conn.Close()
	}


	// TODO test if it's actually alive

	// start new session
	if d.Context != nil {
		d.CancelFunc()
	}
	d.Context, d.CancelFunc = context.WithCancel(context.Background())

	d.commands = make(chan Command)
	go d.commander()

	return nil
}
func (d *Device) Command(cmd Command) {
	d.commands <- cmd
}
func (d *Device) commander() {
	for {
		select {
		case <- d.Done():
			return
		case cmd := <- d.commands:

			// try until success
			for {
				response, err := d.command(cmd)
				if err != nil {
					fmt.Printf("ERROR %T: %v\n", cmd, err)
					time.Sleep(time.Second)
					continue
				}
				go d.handle(response)
				break
			}

		}
	}
}
func (d *Device) command(cmd Command) (response string, err error) {
	fmt.Printf("[Projector.command] %#v\n", cmd)

	var conn net.Conn

	conn, err = net.DialTimeout("tcp", d.Address, time.Second * 1)
	if err != nil {
		return "", fmt.Errorf("DialTCP: %v", err)
	} else {
		defer conn.Close()
	}

	// Write
	err = conn.SetWriteDeadline(time.Now().Add(time.Second))
	if err != nil {
		return "", fmt.Errorf("ERROR setWriteDeadline: %v\n", err)
	}
	_, err = conn.Write([]byte(cmd.Request()))
	if err != nil {
		return "", fmt.Errorf("ERROR conn.Write: %v\n", err)
	}

	// Read
	err = conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	if err != nil {
		return "", fmt.Errorf("ERROR setReadDeadline: %v\n", err)
	}
	data := make([]byte, 100)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Printf("ERROR conn.Read: %v\n", err)
		return
	}
	data = data[:n]

	return string(data), nil
}


func (d *Device) handle(response string) {
	//fmt.Printf("[Projectorhandle] %#v\n", d)

	// TODO detect command type better
	for _, prefix := range d.State.Power.Prefixes() {
		if strings.HasPrefix(response, prefix) {
			d.State.Power.Handle(response)
			d.callback(d.State.Power)
			return
		}
	}
	for _, prefix := range d.State.Freeze.Prefixes() {
		if strings.HasPrefix(response, prefix) {
			d.State.Freeze.Handle(response)
			d.callback(d.State.Freeze)
			return
		}
	}
	for _, prefix := range d.State.Blank.Prefixes() {
		if strings.HasPrefix(response, prefix) {
			d.State.Blank.Handle(response)
			d.callback(d.State.Blank)
			return
		}
	}

	fmt.Printf("UNHANDLED project.handle(%s)\n", response)
}
func (d *Device) callback(cmd interface{}) {
	for _, callback := range d.EventCallbacks {
		callback <- cmd
	}
}