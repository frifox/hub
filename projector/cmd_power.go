package projector

import (
	"strings"
)

type Power struct {
	Status string
}
func (c *Power) Prefixes() []string {
	return []string{
		"*pow=?#",
		"*pow=on#",
		"*pow=off#",
	}
}
func (c *Power) Request() string {
	switch c.Status {
	case "on":
		return "*pow=on#"
	case "off":
		return "*pow=off#"
	default:
		return "*pow=?#"
	}
}
func (c *Power) Handle(line string) {
	for _, prefix := range c.Prefixes() {
		line = strings.TrimPrefix(line, prefix)
	}
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSuffix(line, "#")

	switch line {
	case "POW=ON":
		c.Status = "on"
	case "POW=OFF":
		c.Status = "off"
	case "???":
		c.Status = "on"
	default:
		c.Status = "//" + line
	}
}