package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/godbus/dbus/v5"
)

const MYNAME string = "ch0mler.slugbug"

func closeConnection(conn *dbus.Conn, logger *slog.Logger) {
	logger.Debug("Closing connection", slog.Any("names", conn.Names()))
	conn.Close()
}

// connect to the appropriate DBus bus
func connectToBus(system bool, private bool, logger *slog.Logger) *dbus.Conn {
	var conn *dbus.Conn
	var err error

	if system {
		if IsRootUser() {
			if private {
				conn, err = dbus.SystemBusPrivate()
			} else {
				conn, err = dbus.SystemBus()
			}
			// TODO: close connection if an error occurs
		} else {
			log.Fatalf("Elevated privileges are needed to spy on the system bus")
		}
	} else {
		if IsRootUser() {
			log.Fatalf("Root users do not have access to a per-user session bus. Use -system instead")
		} else {
			if private {
				conn, err = dbus.SessionBusPrivate()
			} else {
				conn, err = dbus.SessionBus()
			}
		}
	}

	if err != nil {
		LogFatal("Failed to connect to dbus", err, logger)
	}

	if private {
		if err = conn.Auth(nil); err != nil {
			LogFatal("Failed to auth to dbus", err, logger)
		}

		if err = conn.Hello(); err != nil {
			LogFatal("Dbus connection is not in a friendly mood right now", err, logger)
		}
	}

	if conn.Connected() {
		busPath := string(conn.BusObject().Path())
		connName := conn.Names()[0]
		logger.Debug("Connected to bus", slog.String("unique_id", connName), slog.String("path", busPath))
	}

	if reply, err := conn.RequestName(MYNAME, dbus.NameFlagReplaceExisting); err != nil {
		logger.Warn("Could not request name", slog.String("name", MYNAME))
	} else if reply == dbus.RequestNameReplyPrimaryOwner {
		logger.Debug(fmt.Sprintf("Successfully bound connection to '%s'", MYNAME), slog.Any("names", conn.Names()))
	} else {
		logger.Warn("Unexpected response when requesting name", slog.String("name", MYNAME), slog.String("reply", reply.String()))
	}

	return conn
}

// initialize DBus monitoring signals
func enableWatch(conn *dbus.Conn, logger *slog.Logger) {
	// connect to signal channel
	ch := make(chan *dbus.Signal, 64)
	conn.Signal(ch)

	busNames := strings.Join(conn.Names(), ",")
	logger.Debug("Watching bus", "names", busNames)

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
			if err := printSignal(s); err != nil {
				logger.Error("Error handling signal", slog.String("error", err.Error()))
			}
		case sig := <-sigc:
			logger.Info("Received signal. Exiting", slog.String("signal", sig.String()))
			break loop
		}
	}
}

// list service objects available to call on the bus
func listBusServices(conn *dbus.Conn, include_unique bool, logger *slog.Logger) []string {
	var (
		listNames     []string
		filteredNames []string
	)
	err := conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&listNames)
	if err != nil {
		LogFatal("Could not list service names", err, logger)
	}

	for _, name := range listNames {
		// own bus name encountered - skip it
		if slices.Contains(conn.Names(), name) {
			logger.Debug("Skipping own bus name", slog.String("name", name))
			continue
		}
		// optionally include unique bus connection names like :1.0
		if !strings.HasPrefix(name, ":") || include_unique {
			filteredNames = append(filteredNames, name)
		}
	}

	logger.Debug(fmt.Sprintf("Filtered %d connections down to %d", len(listNames), len(filteredNames)))
	return filteredNames
}

func printSignal(s *dbus.Signal) error {
	fmt.Printf("%v\n", s)
	return nil
}
