package main

import (
	"fmt"
	"main/decoder"
	"main/encoder"
	"main/projector"
	"strings"
)

type Hub struct {
	Projectors map[string]*projector.Device
	Encoders map[string]*encoder.Device
	Decoders map[string]*decoder.Device
}

func (h *Hub) Add(device string, name string, vars map[string]string) {
	device = strings.ToLower(device)
	name = strings.ToLower(name)

	switch device {
	case "projector":
		dev := &projector.Device{
			Address: vars["address"],
			Name: name,
		}
		go dev.Run()

		if h.Projectors == nil {
			h.Projectors = make(map[string]*projector.Device)
		}
		h.Projectors[name] = dev
	case "encoder":
		dev := &encoder.Device{
			Address: vars["address"],
			Name: vars["name"],
			Channel: vars["channel"],
		}
		go dev.Run()

		if h.Encoders == nil {
			h.Encoders = make(map[string]*encoder.Device)
		}
		h.Encoders[name] = dev
	case "decoder":
		dev := &decoder.Device{
			Address: vars["address"],
		}
		go dev.Run()

		if h.Decoders == nil {
			h.Decoders = make(map[string]*decoder.Device)
		}
		h.Decoders[name] = dev
	default:
		fmt.Printf("hub.Add: unknown device type %s\n", device)
	}
}

func (h *Hub) RunAll() {
	for _, dev := range h.Projectors {
		go dev.Run()
	}
	for _, dev := range h.Encoders {
		go dev.Run()
	}
	for _, dev := range h.Decoders {
		go dev.Run()
	}
}

func (h *Hub) GetProjector(name string) *projector.Device {
	return h.Projectors[name]
}

func (h *Hub) GetEncoder(name string) *encoder.Device {
	return h.Encoders[name]
}

func (h *Hub) GetDecoder(name string) *decoder.Device {
	return h.Decoders[name]
}