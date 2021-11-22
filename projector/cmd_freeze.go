package projector

import (
	"strings"
)

type Freeze struct {
	Status string
}
func (c *Freeze) Prefixes() []string {
	return []string{
		"*freeze=?#",
		"*freeze=on#",
		"*freeze=off#",
	}
}
func (c *Freeze) Request() string {
	switch c.Status {
	case "on":
		return "*freeze=on#"
	case "off":
		return "*freeze=off#"
	default:
		return "*freeze=?#"
	}
}
func (c *Freeze) Handle(line string) {
	for _, prefix := range c.Prefixes() {
		line = strings.TrimPrefix(line, prefix)
	}
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSuffix(line, "#")

	switch line {
	case "FREEZE=ON":
		c.Status = "on"
	case "FREEZE=OFF":
		c.Status = "off"
	case "???":
		c.Status = "off"
	default:
		c.Status = "//" + line
	}
}