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

// some services are commonly returned during introspection
// but we don't want to offer them to the user every time
var SkipInspectionServices = []string{
	"org.freedesktop.DBus.Introspectable",
	"org.freedesktop.DBus.Properties",
	"org.freedesktop.DBus.Peer",
}
