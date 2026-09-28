package slugbug

import (
	"fmt"
	"log/slog"
	"slugbug/internal/helpers"

	"github.com/godbus/dbus/v5"
)

// unique bool, debug bool, system bool, private bool
type Slugbug struct {
	name        string
	conn        *dbus.Conn
	logger      *slog.Logger
	services    []string
	listUnique  bool
	debug       bool
	systemBus   bool
	privateConn bool
}

func NewSlugBug(system bool, private bool, debug bool, unique bool) *Slugbug {
	s := &Slugbug{
		name:        "ch0mler.slugbug",
		conn:        nil,
		logger:      helpers.InitLogging(debug),
		debug:       debug,
		systemBus:   system,
		listUnique:  unique,
		privateConn: private,
	}
	return s
}

func (s *Slugbug) Name() string {
	return s.name
}

func (s *Slugbug) Conn() *dbus.Conn {
	return s.conn
}

// connect to the appropriate DBus bus
func (s *Slugbug) ConnectToBus() error {
	var err error

	if s.systemBus {
		if helpers.IsRootUser() {
			if s.privateConn {
				s.conn, err = dbus.SystemBusPrivate()
			} else {
				s.conn, err = dbus.SystemBus()
			}
		} else {
			return fmt.Errorf("elevated privileges are needed to spy on the system bus")
		}
	} else {
		if helpers.IsRootUser() {
			return fmt.Errorf("root users do not have access to a per-user session bus; use -system instead")
		} else {
			if s.privateConn {
				s.conn, err = dbus.SessionBusPrivate()
			} else {
				s.conn, err = dbus.SessionBus()
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to dbus: %w", err)
	}

	if s.privateConn {
		if err = s.conn.Auth(nil); err != nil {
			s.conn.Close()
			s.conn = nil
			return fmt.Errorf("failed to auth to dbus: %w", err)
		}

		if err = s.conn.Hello(); err != nil {
			s.conn.Close()
			s.conn = nil
			return fmt.Errorf("dbus connection is not in a friendly mood right now: %w", err)
		}
	}

	if s.conn.Connected() {
		busPath := string(s.conn.BusObject().Path())
		connName := s.conn.Names()[0]
		s.logger.Debug("Connected to bus", slog.String("unique_id", connName), slog.String("path", busPath))
	}

	var reply dbus.RequestNameReply

	if reply, err = s.conn.RequestName(s.name, dbus.NameFlagReplaceExisting); err != nil {
		return fmt.Errorf("could not request name %q: %w", s.name, err)
	} else if reply == dbus.RequestNameReplyPrimaryOwner {
		s.logger.Debug(fmt.Sprintf("Successfully bound connection to '%s'", s.name), slog.Any("names", s.conn.Names()))
	} else {
		s.logger.Warn("Unexpected response when requesting name", slog.String("name", s.name), slog.String("reply", reply.String()))
	}

	return nil
}

// disconnect from the DBus
func (s *Slugbug) CloseConnection() {
	if s.conn == nil {
		return
	}
	s.logger.Debug("Releasing connection name", slog.String("name", s.name))
	s.conn.ReleaseName(s.name)
	s.logger.Debug("Closing connection", slog.Any("names", s.conn.Names()))
	s.conn.Close()
	s.conn = nil
}

// only call ListServices once to save on processing
func (s *Slugbug) Services() ([]string, error) {
	if len(s.services) == 0 {
		services, err := s.ListBusServices()
		if err != nil {
			return nil, err
		}
		s.services = services
	}
	return s.services, nil
}
