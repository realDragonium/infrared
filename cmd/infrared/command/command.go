package command

type Command string

const (
	Reload Command = "reload"
)

var commands = map[Command][]interface{}{
	Reload: nil,
}
