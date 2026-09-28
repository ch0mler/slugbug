package slugbug

import (
	"github.com/godbus/dbus/v5"
)

// custom interface made to allow for mocking dbus connection in tests
type busConnection interface {
	// methods which call the underlying dbus connection methods
	Auth() error
	Hello() error
	Connected() bool
	Close() error
	BusObject() dbus.BusObject
	Object(string, dbus.ObjectPath) dbus.BusObject
	BusPath() dbus.ObjectPath
	Names() []string
	RequestName(string, dbus.RequestNameFlags) (dbus.RequestNameReply, error)
	ReleaseName(string) (dbus.ReleaseNameReply, error)
	// methods which are not found on the underlying dbus connection
	ListNames() ([]string, error)
}

type dbusConnection struct {
	conn *dbus.Conn
}

func (c *dbusConnection) Auth() error {
	return c.conn.Auth(nil)
}

func (c *dbusConnection) Hello() error {
	return c.conn.Hello()
}

func (c *dbusConnection) Connected() bool {
	return c.conn.Connected()
}

func (c *dbusConnection) BusObject() dbus.BusObject {
	return c.conn.BusObject()
}

func (c *dbusConnection) BusPath() dbus.ObjectPath {
	return c.BusObject().Path()
}

func (c *dbusConnection) Object(dest string, path dbus.ObjectPath) dbus.BusObject {
	return c.conn.Object(dest, path)
}

func (c *dbusConnection) Names() []string {
	return c.conn.Names()
}

func (c *dbusConnection) RequestName(name string, flags dbus.RequestNameFlags) (dbus.RequestNameReply, error) {
	return c.conn.RequestName(name, flags)
}

func (c *dbusConnection) ReleaseName(name string) (dbus.ReleaseNameReply, error) {
	return c.conn.ReleaseName(name)
}

func (c *dbusConnection) Close() error {
	return c.conn.Close()
}

func (c *dbusConnection) ListNames() ([]string, error) {
	var names []string
	err := c.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	return names, err
}
