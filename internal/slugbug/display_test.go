package slugbug

import (
	"slices"
	"strings"
	"testing"
)

func newSlugbugWithBus(listUnique bool) *Slugbug {
	sb := NewSlugBug(false, false, false, listUnique)
	sb.conn = &fakeBusConnection{
		listedNames: []string{
			"org.freedesktop.DBus",
			"org.example.SlugbugTest",
			":1.41",
			":1.42",
		},
		ownedNames: []string{":1.41"},
	}
	return sb
}

type fakeBusConnection struct {
	busConnection
	listedNames []string
	ownedNames  []string
}

func (c *fakeBusConnection) ListNames() ([]string, error) {
	return c.listedNames, nil
}

func (c *fakeBusConnection) Names() []string {
	return c.ownedNames
}

func TestListBusServicesWithoutConnection(t *testing.T) {
	sb := NewSlugBug(false, false, false, false)

	services, err := sb.ListBusServices()
	if err == nil {
		t.Fatal("ListBusServices() error = nil, want an error")
	}
	if services != nil {
		t.Fatalf("ListBusServices() services = %v, want nil", services)
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Fatalf("ListBusServices() error = %q, want not-connected message", err)
	}
}

func TestListBusServicesReturnsServiceNames(t *testing.T) {
	sb := newSlugbugWithBus(false)

	services, err := sb.ListBusServices()
	if err != nil {
		t.Fatalf("ListBusServices() error = %v", err)
	}

	slices.Sort(services)
	want := []string{"org.example.SlugbugTest", "org.freedesktop.DBus"}
	if !slices.Equal(services, want) {
		t.Fatalf("ListBusServices() = %v, want %v", services, want)
	}
}

func TestListBusServicesFiltersUniqueNamesByDefault(t *testing.T) {
	sb := newSlugbugWithBus(false)

	services, err := sb.ListBusServices()
	if err != nil {
		t.Fatalf("ListBusServices() error = %v", err)
	}

	for _, service := range services {
		if strings.HasPrefix(service, ":") {
			t.Errorf("ListBusServices() returned unique name %q by default", service)
		}
	}
}

func TestListBusServicesIncludesUniqueNamesWhenRequested(t *testing.T) {
	sb := newSlugbugWithBus(true)

	services, err := sb.ListBusServices()
	if err != nil {
		t.Fatalf("ListBusServices() error = %v", err)
	}

	uniqueName := ":1.42"
	if !slices.Contains(services, uniqueName) {
		t.Errorf("ListBusServices() = %v, want unique name %q", services, uniqueName)
	}
}
