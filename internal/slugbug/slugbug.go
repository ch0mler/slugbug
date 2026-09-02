package slugbug

import (
	"log/slog"
	"slugbug/internal/utils"

	"github.com/godbus/dbus/v5"
)

type Slugbug struct {
	name   string
	conn   *dbus.Conn
	logger *slog.Logger
}

func NewSlugBug(system bool, private bool, debug bool) *Slugbug {
	s := &Slugbug{
		name:   "ch0mler.slugbug",
		conn:   nil,
		logger: utils.InitLogging(debug),
	}
	s.ConnectToBus(system, private)
	return s
}
