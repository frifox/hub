package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"main/decoder"
	"main/shared"
	"net/http"
	"sort"
	"strings"
)

type Client struct {
	WS *websocket.Conn
	write chan interface{}

	projectorCallbacks map[string]chan interface{}

	context.Context
	context.CancelFunc
}

var wsUpgrade = websocket.Upgrader{}
func (c *Client) Serve(w http.ResponseWriter, r *http.Request) {
	// handle new request
	var err error
	c.WS, err = wsUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	// request is now upgraded to websocket
	c.Context, c.CancelFunc = context.WithCancel(context.Background())

	// handle websocket until we're done
	c.write = make(chan interface{})
	go c.Reader()
	go c.Writer()

	c.projectorCallbacks = make(map[string]chan interface{})
	go c.RegisterCallbacks()
	<- c.Done()

	if err := c.WS.Close(); err != nil {
		log.Printf("ERR ws close(): %v\n", err)
	}
}

type Emit struct {
	DeviceType string
	DeviceName string
	Command string
	Data interface{} `json:",omitempty"`
}
func (c *Client) Reader() {
	for {
		dataType, data, err := c.WS.ReadMessage()
		if err != nil {
			fmt.Printf("ERROR ws.ReadMessage(): %v\n", err)
			c.CancelFunc()
			return
		}

		fmt.Printf("[FromClient] %s\n", data)

		switch dataType {
		default:
			fmt.Printf("WARNING emitType(%d): %+v\n", dataType, data)
		case websocket.TextMessage:
			emit := Emit{}
			if err := json.Unmarshal(data, &emit); err != nil {
				fmt.Printf("ERROR json.Unmarshal(emit): %v\n", err)
				continue
			}

			switch emit.DeviceType {
			case "projector":
				dev := hub.GetProjector(emit.DeviceName)
				if dev == nil {
					fmt.Printf("ERROR projectors.Get(%s) 404\n", emit.DeviceName)
					continue
				}
				if err := dev.Handle(data); err != nil {
					fmt.Printf("ERROR projector.Handle(%s): %v\n", emit, err)
					continue
				}

			case "decoder":
				dev := hub.GetDecoder(emit.DeviceName)
				if dev == nil {
					fmt.Printf("ERROR decoders.Get(%s) 404\n", emit.DeviceName)
					continue
				}

				// transform incoming emit
				if err := c.TransformDecoderEmit(&data); err != nil {
					fmt.Printf("TransformDecoderEmit: %v\n", err)
					continue
				}

				if err := dev.Handle(data); err != nil {
					fmt.Printf("ERROR decoder.Handle(%s): %v\n", emit, err)
					continue
				}

			default:
				//fmt.Printf("UNHANDLED DeviceType: %v\n", emit.DeviceType)
			}
		}
	}
}

func (c *Client) TransformDecoderEmit(data *[]byte) (err error) {
	emit := decoder.Emit{}
	if err = json.Unmarshal(*data, &emit); err != nil {
		return fmt.Errorf("unmarshal(decoder.Emit): %v", err)
	}

	// inject Encoder info
	if emit.Command == "ConnectTo" {
		// get encoder name
		emit2 := struct{
			Encoder string
		}{}
		if err := json.Unmarshal(*data, &emit2); err != nil {
			return fmt.Errorf("unmarshal(decoder.emit2): %v", err)
		}

		// find it
		enc := hub.GetEncoder(emit2.Encoder)
		if enc == nil {
			return fmt.Errorf("hub.GetEncoder(emit2.Encoder): 404")
		}

		// prepare Decoder-friendly emit
		addr := strings.Split(enc.Address, ":")
		cmd := struct {
			decoder.Emit
			decoder.ConnectTo
		}{
			Emit: decoder.Emit{
				Command: shared.KeyOf(decoder.ConnectTo{}),
			},
			ConnectTo: decoder.ConnectTo{
				ConnectToIp:  addr[0],
				Port:         addr[1],
				SourceName:   enc.Name,
				SourcePCName: enc.Channel,
			},
		}
		bs, err := json.Marshal(&cmd)
		if err != nil {
			return fmt.Errorf("json.Marshal(decoder.ConnectTo): %v\n", err)
		}

		// apply Decoder-friendly data
		*data = bs
	}
	return nil
}

func (c *Client) Writer() {
	for {
		select {
		case emit := <-c.write:
			fmt.Printf("[ToClient] %#v\n", emit)
			data, err := json.Marshal(emit)
			if err != nil {
				fmt.Printf("ERROR json.Marshal(%+v)\n", emit)
				continue
			}
			err = c.WS.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Printf("ERROR WS.WriteJSON: %v\n", err)
				continue
			}
		case <- c.Done():
			close(c.write)
			return
		}
	}
}

func (c *Client) RegisterCallbacks() {
	clientID := fmt.Sprintf("%p", c)

	// prepare receiving channels
	for name, _ := range hub.Projectors {
		c.projectorCallbacks[name] = make(chan interface{})
	}

	// send device list to client
	c.EmitDeviceList()

	// register callbacks with devices
	for name, dev := range hub.Projectors {
		go c.ReadCallbacks("projector", name, c.projectorCallbacks[name])
		dev.Register(clientID, c.projectorCallbacks[name])
	}

	// wait for client to disconnect
	<- c.Done()

	// deregister
	for _, dev := range hub.Projectors {
		dev.Deregister(clientID)
	}
}
func (c *Client) ReadCallbacks(DeviceType string, DeviceName string, channel chan interface{}) {
	for {
		select {
		case cmd := <- channel:
			fmt.Printf("[ClientCallback] %+v\n", cmd)
			emit := &Emit{
				DeviceType: DeviceType,
				DeviceName: DeviceName,
				Command: shared.KeyOf(cmd),
				Data: cmd,
			}
			c.write <- emit
		case <- c.Done():

			close(channel)
			return
		}
	}
}
func (c *Client) EmitDeviceList() {
	var list []string

	// emit projectors
	list = []string{}
	for name, _ := range hub.Projectors {
		list = append(list, name)
	}
	sort.Strings(list) // consistency is key
	for _, name := range list {
		emit := &Emit{
			DeviceType: "projector",
			DeviceName: name,
			Command: "Init",
		}
		c.write <- emit
	}

	// emit decoders
	list = []string{}
	for name, _ := range hub.Decoders {
		list = append(list, name)
	}
	sort.Strings(list) // consistency is key
	for _, name := range list {
		emit := &Emit{
			DeviceType: "decoder",
			DeviceName: name,
			Command: "Init",
		}
		c.write <- emit
	}

	// emit encoders
	list = []string{}
	for name, _ := range hub.Encoders {
		list = append(list, name)
	}
	sort.Strings(list) // consistency is key
	for _, name := range list {
		emit := &Emit{
			DeviceType: "encoder",
			DeviceName: name,
			Command: "Init",
		}
		c.write <- emit
	}
}