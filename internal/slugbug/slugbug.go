package slugbug

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"slugbug/internal/helpers"
	"strings"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
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

func (s *Slugbug) InspectService(service string) (string, error) {
	svcObj := s.conn.Object(service, s.conn.BusObject().Path())
	node, err := introspect.Call(svcObj)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	for _, v := range node.Interfaces {
		if slices.Contains(SkipInspectionServices, v.Name) {
			continue
		}
		fmt.Fprintln(&result, v.Name)
		writeAnnotations(&result, v.Annotations)
		writeMethods(&result, v.Methods)
		writeProperties(&result, v.Properties)
		writeSignals(&result, v.Signals)
	}
	return result.String(), nil
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

func writeAnnotations(result *strings.Builder, annotations []introspect.Annotation) {
	if len(annotations) == 0 {
		return
	}
	fmt.Fprintln(result, "  Annotations")
	fmt.Fprintf(result, "    > %s\n", formatAnnotations(annotations))
}

func writeMethods(result *strings.Builder, methods []introspect.Method) {
	if len(methods) == 0 {
		return
	}
	fmt.Fprintln(result, "  Methods")
	for _, method := range methods {
		fmt.Fprintf(result, "    > %s%s\n", method.Name, formatArgs(method.Args))
	}
}

func writeProperties(result *strings.Builder, properties []introspect.Property) {
	if len(properties) == 0 {
		return
	}
	fmt.Fprintln(result, "  Properties")
	for _, property := range properties {
		fmt.Fprintf(result, "    > Name: %s\n", property.Name)
		fmt.Fprintf(result, "      Type: %s\n", property.Type)
		fmt.Fprintf(result, "      Access: %s\n", property.Access)
		fmt.Fprintf(result, "      Annotations: %s\n", formatAnnotations(property.Annotations))
	}
}

func writeSignals(result *strings.Builder, signals []introspect.Signal) {
	if len(signals) == 0 {
		return
	}
	fmt.Fprintln(result, "  Signals")
	for _, signal := range signals {
		fmt.Fprintf(result, "    > %s%s\n", signal.Name, formatArgs(signal.Args))
		if len(signal.Annotations) > 0 {
			fmt.Fprintf(result, "    > Annotations: %s\n", formatAnnotations(signal.Annotations))
		}
	}
}

func formatAnnotations(annotations []introspect.Annotation) string {
	if len(annotations) == 0 {
		return ""
	}
	var ret strings.Builder

	for _, annotation := range annotations {
		fmt.Fprintf(&ret, "%s %s", annotation.Value, annotation.Name)
	}

	return ret.String()
}

func formatArgs(args []introspect.Arg) string {
	var (
		inArgs  []string
		outArgs []string
		ret     strings.Builder
	)

	// types and parameters dictate output
	for _, arg := range args {
		switch arg.Direction {
		case "in":
			if arg.Name == "" {
				inArgs = append(inArgs, string(arg.Type))
			} else {
				inArgs = append(inArgs, fmt.Sprintf("%s %s", arg.Name, arg.Type))
			}
		case "out":
			if arg.Name == "" {
				outArgs = append(outArgs, string(arg.Type))
			} else {
				outArgs = append(outArgs, fmt.Sprintf("%s %s", arg.Name, arg.Type))
			}
		}
	}

	// format the arguments like a method call in other languages
	if len(inArgs) > 0 && len(outArgs) > 0 {
		fmt.Fprintf(&ret, "(%s) -> %s", strings.Join(inArgs, ", "), strings.Join(outArgs, ", "))
	} else if len(inArgs) > 0 {
		fmt.Fprintf(&ret, "(%s)", strings.Join(inArgs, ", "))
	} else if len(outArgs) > 0 {
		fmt.Fprintf(&ret, "() -> %s", strings.Join(outArgs, ", "))
	}

	return ret.String()
}
