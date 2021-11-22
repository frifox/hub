package decoder

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var SupportedCommands = []Command{
	&List{},
	&ConnectTo{},
}

type List struct {}
func (c *List) Request() *http.Request {
	r, err := http.NewRequest("GET", "/List", nil)
	if err != nil {
		panic("bad /List request:" + err.Error())
	}

	return r
}

type ConnectTo struct {
	ConnectToIp string `json:"connectToIp"`
	Port string `json:"port"`
	SourceName string `json:"sourceName"`
	SourcePCName string `json:"sourcePcName"`
}
func (c *ConnectTo) Request() *http.Request {
	body, err := json.Marshal(c)
	if err != nil {
		panic("bad json.Marshal(ConnectTo): " + err.Error())
	}

	r, err := http.NewRequest("POST", "/connectTo", bytes.NewReader(body))
	if err != nil {
		panic("bad /connectTo request:" + err.Error())
	}

	return r
}