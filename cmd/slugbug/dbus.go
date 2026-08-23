package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"
)

const MYNAME string = "ch0mler.slugbug"

// connect to the appropriate DBus bus
func connectToBus(system bool, private bool) *dbus.Conn {
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
		fmt.Printf("Connected to %s\n", conn.BusObject().Path())
	}

	if reply, err := conn.RequestName(MYNAME, dbus.NameFlagReplaceExisting); err != nil {
		log.Printf("Could not request name: '%s'\n", MYNAME)
	} else if reply == dbus.RequestNameReplyPrimaryOwner {
		log.Printf("Successfully bound connection to name: %s\n", MYNAME)
	} else {
		log.Printf("Unexpected response when requesting name '%s': %s\n", MYNAME, reply.String())
	}

	return conn
}

// initialize DBus monitoring signals
func enableWatch(conn *dbus.Conn) {
	// connect to signal channel
	ch := make(chan *dbus.Signal, 64)
	conn.Signal(ch)

	fmt.Printf("%v\n", conn.Names())
	// log.Printf("Monitoring the system bus\n")
	// log.Print("Monitoring the session bus\n")

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
				log.Printf("Error handling signal: %v", err)
			}
		case sig := <-sigc:
			log.Printf("Received signal %s, exiting", sig)
			break loop
		}
	}
}

func printSignal(s *dbus.Signal) error {
	fmt.Printf("%v\n", s)
	return nil
}
