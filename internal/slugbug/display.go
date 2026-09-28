package slugbug

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

const Display Operation = "Display available services"

// list service objects available to call on the bus
func (s *Slugbug) ListBusServices() ([]string, error) {
	var (
		listNames     []string
		filteredNames []string
	)

	if s.conn == nil {
		return nil, fmt.Errorf("not connected to dbus")
	}
	if err := s.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&listNames); err != nil {
		return nil, fmt.Errorf("could not list service names: %w", err)
	}

	for _, name := range listNames {
		// own bus name encountered - skip it
		if slices.Contains(s.conn.Names(), name) {
			s.logger.Debug("Skipping own bus name", slog.String("name", name))
			continue
		}
		// optionally include unique bus connection names like :1.0
		if !strings.HasPrefix(name, ":") || s.listUnique {
			filteredNames = append(filteredNames, name)
		}
	}

	s.logger.Debug(fmt.Sprintf("Filtered %d connections down to %d", len(listNames), len(filteredNames)))
	return filteredNames, nil
}
