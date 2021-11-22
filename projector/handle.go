package projector

import (
	"encoding/json"
	"fmt"
	"main/shared"
	"strings"
)

type ClientEmit struct {
	Command string
}

func (d *Device) Handle(data []byte) (err error) {
	fmt.Printf("[Device.Handle] %s\n", data)

	emit := ClientEmit{}
	if err := json.Unmarshal(data, &emit); err != nil {
		return fmt.Errorf("json.Unmarshal(): %v\n", err)
	}

	// check all statuses?
	if emit.Command == "Refresh" {
		for _, cmd := range SupportedCommands {
			go d.Command(cmd)
		}
		return nil
	}

	// one-off command?
	for _, cmd := range SupportedCommands {
		if strings.EqualFold(shared.KeyOf(cmd), emit.Command) {
			if err := json.Unmarshal(data, &cmd); err != nil {
				fmt.Printf("json.Unmarshal(cmd): %v\n", err)
				continue
			}
			go d.Command(cmd)
			return nil
		}
	}

	return fmt.Errorf("unhandled command")
}
