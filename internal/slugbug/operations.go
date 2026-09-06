package slugbug

type Operation string

const (
	Display Operation = "Display available services"
	Call    Operation = "Call a method"
	Inspect Operation = "Inspect a service"
	Monitor Operation = "Monitor signal(s)"
	Quit    Operation = "Quit"
)

func (o Operation) String() string {
	return string(o)
}
