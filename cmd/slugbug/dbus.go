package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/godbus/dbus/v5"
)

const MYNAME string = "ch0mler.slugbug"

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
		} else {
			log.Fatal("Elevated privileges are needed to spy on the system bus")
		}
	} else {
		if IsRootUser() {
			log.Fatalf("Root users do not have access to a per-user session bus. Use --system instead")
		} else {
			if private {
				conn, err = dbus.SessionBusPrivate()
			} else {
				conn, err = dbus.SessionBus()
			}
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to dbus: %v", err)
	}
	defer conn.Close()

	if private {
		if err = conn.Auth(nil); err != nil {
			log.Fatalf("Failed to auth to dbus: %v", err)
		}

		if err = conn.Hello(); err != nil {
			log.Fatalf("Dbus connection is not in a friendly mood right now: %v", err)
		}
	}

	if conn.Connected() {
		busPath := string(conn.BusObject().Path())
		logger.Info("Connected to bus", slog.String("path", busPath))
	}

	if reply, err := conn.RequestName(MYNAME, dbus.NameFlagReplaceExisting); err != nil {
		logger.Warn("Could not request name", slog.String("name", MYNAME))
	} else if reply == dbus.RequestNameReplyPrimaryOwner {
		logger.Debug("Successfully bound connection to name", slog.String("name", MYNAME))
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

func printSignal(s *dbus.Signal) error {
	fmt.Printf("%v\n", s)
	return nil
}
