package decoder

import (
	"encoding/json"
	"fmt"
	"main/shared"
	"strings"
)

type Emit struct {
	Command string
}

func (d *Device) Handle(data []byte) (err error) {
	emit := Emit{}
	if err := json.Unmarshal(data, &emit); err != nil {
		return fmt.Errorf("json.Unmarshal: %v", err)
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

	return fmt.Errorf("unhandled emit(%s)", data)
}