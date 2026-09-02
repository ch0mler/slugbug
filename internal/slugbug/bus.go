package slugbug

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"slugbug/internal/utils"
	"strings"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

// connect to the appropriate DBus bus
func (s *Slugbug) ConnectToBus(system bool, private bool) {
	var err error

	if system {
		if utils.IsRootUser() {
			if private {
				s.conn, err = dbus.SystemBusPrivate()
			} else {
				s.conn, err = dbus.SystemBus()
			}
			// TODO: close connection if an error occurs
		} else {
			log.Fatalf("Elevated privileges are needed to spy on the system bus")
		}
	} else {
		if utils.IsRootUser() {
			log.Fatalf("Root users do not have access to a per-user session bus. Use -system instead")
		} else {
			if private {
				s.conn, err = dbus.SessionBusPrivate()
			} else {
				s.conn, err = dbus.SessionBus()
			}
		}
	}

	if err != nil {
		utils.LogFatal("Failed to connect to dbus", err, s.logger)
	}

	if private {
		if err = s.conn.Auth(nil); err != nil {
			utils.LogFatal("Failed to auth to dbus", err, s.logger)
		}

		if err = s.conn.Hello(); err != nil {
			utils.LogFatal("Dbus connection is not in a friendly mood right now", err, s.logger)
		}
	}

	if s.conn.Connected() {
		busPath := string(s.conn.BusObject().Path())
		connName := s.conn.Names()[0]
		s.logger.Debug("Connected to bus", slog.String("unique_id", connName), slog.String("path", busPath))
	}

	if reply, err := s.conn.RequestName(s.name, dbus.NameFlagReplaceExisting); err != nil {
		s.logger.Warn("Could not request name", slog.String("name", s.name))
	} else if reply == dbus.RequestNameReplyPrimaryOwner {
		s.logger.Debug(fmt.Sprintf("Successfully bound connection to '%s'", s.name), slog.Any("names", s.conn.Names()))
	} else {
		s.logger.Warn("Unexpected response when requesting name", slog.String("name", s.name), slog.String("reply", reply.String()))
	}
}

// disconnect from the DBus
func (s *Slugbug) CloseConnection() {
	s.logger.Debug("Closing connection", slog.Any("names", s.conn.Names()))
	s.conn.Close()
}

// initialize DBus monitoring signals
func (s *Slugbug) EnableWatch() {
	// connect to signal channel
	ch := make(chan *dbus.Signal, 64)
	s.conn.Signal(ch)

	busNames := strings.Join(s.conn.Names(), ",")
	s.logger.Debug("Watching bus", "names", busNames)

	// Handle termination signals to clean up
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
loop:
	for {
		select {
		case s := <-ch:
			if s == nil {
				break loop
			}
		case sig := <-sigc:
			s.logger.Info("Received signal. Exiting", slog.String("signal", sig.String()))
			break loop
		}
	}
}

func (s *Slugbug) InspectBus(service string) {
	node, err := introspect.Call(s.conn.Object(service, s.conn.BusObject().Path()))
	if err != nil {
		log.Fatal(err)
	}
	// nodeDetails := introspect.NewIntrospectable(node)
	// fmt.Println(nodeDetails.Introspect())
	for _, v := range node.Interfaces {
		fmt.Printf("Signals for %s\n", v.Name)
		for _, k := range v.Signals {
			fmt.Printf("%v\n", k)
		}
	}
}

// list service objects available to call on the bus
func (s *Slugbug) ListBusServices(include_unique bool) []string {
	var (
		listNames     []string
		filteredNames []string
	)
	err := s.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&listNames)
	if err != nil {
		utils.LogFatal("Could not list service names", err, s.logger)
	}

	for _, name := range listNames {
		// own bus name encountered - skip it
		if slices.Contains(s.conn.Names(), name) {
			s.logger.Debug("Skipping own bus name", slog.String("name", name))
			continue
		}
		// optionally include unique bus connection names like :1.0
		if !strings.HasPrefix(name, ":") || include_unique {
			filteredNames = append(filteredNames, name)
		}
	}

	s.logger.Debug(fmt.Sprintf("Filtered %d connections down to %d", len(listNames), len(filteredNames)))
	return filteredNames
}
