package projector

import (
	"strings"
)

type Blank struct {
	Status string
}
func (c *Blank) Prefixes() []string {
	return []string{
		"*blank=?#",
		"*blank=on#",
		"*blank=off#",
	}
}
func (c *Blank) Request() string {
	switch c.Status {
	case "on":
		return "*blank=on#"
	case "off":
		return "*blank=off#"
	default:
		return "*blank=?#"
	}
}
func (c *Blank) Handle(line string) {
	for _, prefix := range c.Prefixes() {
		line = strings.TrimPrefix(line, prefix)
	}
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSuffix(line, "#")

	switch line {
	case "BLANK=ON":
		c.Status = "on"
	case "BLANK=OFF":
		c.Status = "off"
	case "???":
		c.Status = "off"
	default:
		c.Status = "//" + line
	}
}