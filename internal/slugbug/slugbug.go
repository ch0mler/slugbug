package slugbug

import (
	"fmt"
	"log"
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

var err error

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
func (s *Slugbug) ConnectToBus() {
	if s.systemBus {
		if helpers.IsRootUser() {
			if s.privateConn {
				s.conn, err = dbus.SystemBusPrivate()
			} else {
				s.conn, err = dbus.SystemBus()
			}
		} else {
			log.Fatalf("Elevated privileges are needed to spy on the system bus")
		}
	} else {
		if helpers.IsRootUser() {
			log.Fatalf("Root users do not have access to a per-user session bus. Use -system instead")
		} else {
			if s.privateConn {
				s.conn, err = dbus.SessionBusPrivate()
			} else {
				s.conn, err = dbus.SessionBus()
			}
		}
	}

	if err != nil {
		helpers.LogFatal("Failed to connect to dbus", err, s.logger)
	}

	if s.privateConn {
		if err = s.conn.Auth(nil); err != nil {
			helpers.LogFatal("Failed to auth to dbus", err, s.logger)
		}

		if err = s.conn.Hello(); err != nil {
			helpers.LogFatal("Dbus connection is not in a friendly mood right now", err, s.logger)
		}
	}

	if s.conn.Connected() {
		busPath := string(s.conn.BusObject().Path())
		connName := s.conn.Names()[0]
		s.logger.Debug("Connected to bus", slog.String("unique_id", connName), slog.String("path", busPath))
	}

	var reply dbus.RequestNameReply

	if reply, err = s.conn.RequestName(s.name, dbus.NameFlagReplaceExisting); err != nil {
		s.logger.Warn("Could not request name", slog.String("name", s.name))
	} else if reply == dbus.RequestNameReplyPrimaryOwner {
		s.logger.Debug(fmt.Sprintf("Successfully bound connection to '%s'", s.name), slog.Any("names", s.conn.Names()))
	} else {
		s.logger.Warn("Unexpected response when requesting name", slog.String("name", s.name), slog.String("reply", reply.String()))
	}
}

// disconnect from the DBus
func (s *Slugbug) CloseConnection() {
	s.logger.Debug("Releasing connection name", slog.String("name", s.name))
	s.conn.ReleaseName(s.name)
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

// list service objects available to call on the bus
func (s *Slugbug) ListBusServices() []string {
	var (
		listNames     []string
		filteredNames []string
	)

	if err = s.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&listNames); err != nil {
		helpers.LogFatal("Could not list service names", err, s.logger)
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
	return filteredNames
}

func (s *Slugbug) InspectService(service string) {
	var node *introspect.Node

	svcObj := s.conn.Object(service, s.conn.BusObject().Path())
	if node, err = introspect.Call(svcObj); err != nil {
		log.Fatal(err)
	}
	for _, v := range node.Interfaces {
		if slices.Contains(SkipInspectionServices, v.Name) {
			continue
		}
		fmt.Println(v.Name)
		printAnnotations(v.Annotations)
		printMethods(v.Methods)
		printProperties(v.Properties)
		printSignals(v.Signals)
	}
}

// only call ListServices once to save on processing
func (s *Slugbug) Services() []string {
	if len(s.services) == 0 {
		s.services = s.ListBusServices()
	}
	return s.services
}

func printAnnotations(annotations []introspect.Annotation) {
	if len(annotations) == 0 {
		return
	}
	fmt.Printf("  Annotations\n")
	fmt.Printf("    > %s", formatAnnotations(annotations))
}

func printMethods(methods []introspect.Method) {
	if len(methods) == 0 {
		return
	}
	fmt.Printf("  Methods\n")
	for _, method := range methods {
		fmt.Printf("    > %s%s\n", method.Name, formatArgs(method.Args))
	}
}

func printProperties(properties []introspect.Property) {
	if len(properties) == 0 {
		return
	}
	fmt.Printf("  Properties\n")
	for _, property := range properties {
		fmt.Printf("    > Name: %s\n", property.Name)
		fmt.Printf("      Type: %s\n", property.Type)
		fmt.Printf("      Access: %s\n", property.Access)
		fmt.Printf("      Annotations: %s\n", formatAnnotations(property.Annotations))
	}
}

func printSignals(signals []introspect.Signal) {
	if len(signals) == 0 {
		return
	}
	fmt.Printf("  Signals\n")
	for _, signal := range signals {
		fmt.Printf("    > %s%s\n", signal.Name, formatArgs(signal.Args))
		if len(signal.Annotations) > 0 {
			fmt.Printf("    > Annotations: %s\n", formatAnnotations(signal.Annotations))
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
