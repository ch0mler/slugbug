package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"
)

// connect to the appropriate DBus bus
func connectToBus(system bool) *dbus.Conn {
	var conn *dbus.Conn
	var err error

	if system {
		if IsRootUser() {
			conn, err = dbus.SystemBus()
		} else {
			log.Fatal("Elevated privileges are needed to spy on the system bus")
		}
	} else {
		conn, err = dbus.SessionBus()
	}

	if err != nil {
		log.Fatalf("failed to connect to dbus: %v", err)
	}
	defer conn.Close()

	if conn.Connected() {
		fmt.Printf("Connected to %s\n", conn.BusObject().Path())
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
