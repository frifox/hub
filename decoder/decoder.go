package decoder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Devices struct {
	Desk Device
}

type Device struct {
	Address string

	Client http.Client
	context.Context
	context.CancelFunc
}

type Command interface {
	Request() *http.Request
}

func (d *Device) Run() {

	for {
		if err := d.Connect(); err != nil {
			fmt.Printf("[%T] ERROR Connect: %v\n", d, err)
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

func (d *Device) Connect() (err error) {
	d.Client = http.Client{}

	// test if it's actually alive
	//cmd := &List{}
	//body, err := d.Command(cmd)
	//if err != nil {
	//	return fmt.Errorf("/List err: %w", err)
	//}

	//fmt.Printf("[%T] %s\n", cmd, body)

	// kill off old stuff, if alive
	if d.Context != nil {
		d.CancelFunc()
	}

	// start new session
	d.Context, d.CancelFunc = context.WithCancel(context.Background())

	return nil
}
func (d *Device) Command(c Command) {
	req, err := http.NewRequest(
		c.Request().Method,
		d.Address + c.Request().URL.String(),
		c.Request().Body,
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.Client.Do(req)
	if err != nil {
		fmt.Printf("request failed: %v", err)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read request failed: %v", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("resp != 200: %d", resp.StatusCode)
		return
	}

	fmt.Printf("Decoder Response: %s\n", data)

	return
}

