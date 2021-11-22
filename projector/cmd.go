package projector

type Command interface {
	Prefixes() []string
	Request() string
	Handle(string)
}

var SupportedCommands = []Command{
	&Freeze{},
	&Power{},
	&Blank{},
}