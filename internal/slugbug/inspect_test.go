package slugbug

import (
	"strings"
	"testing"

	"github.com/godbus/dbus/v5/introspect"
)

func TestFormatAnnotations(t *testing.T) {
	annotations := []introspect.Annotation{
		{Name: "org.example.First", Value: "one"},
		{Name: "org.example.Second", Value: "two"},
	}

	got := formatAnnotations(annotations)
	want := "one org.example.Firsttwo org.example.Second"
	if got != want {
		t.Fatalf("formatAnnotations() = %q, want %q", got, want)
	}
}

func TestFormatArgs(t *testing.T) {
	args := []introspect.Arg{
		{Name: "message", Type: "s", Direction: "in"},
		{Type: "u", Direction: "out"},
	}

	got := formatArgs(args)
	want := "(message s) -> u"
	if got != want {
		t.Fatalf("formatArgs() = %q, want %q", got, want)
	}
}

func TestWriteAnnotations(t *testing.T) {
	var result strings.Builder
	writeAnnotations(&result, []introspect.Annotation{{Name: "org.example.Foo", Value: "bar"}})

	got := result.String()
	if !strings.Contains(got, "Annotations") || !strings.Contains(got, "bar org.example.Foo") {
		t.Fatalf("writeAnnotations() = %q, want annotation header and value", got)
	}
}

func TestWriteMethods(t *testing.T) {
	var result strings.Builder
	writeMethods(&result, []introspect.Method{{
		Name: "Ping",
		Args: []introspect.Arg{{Name: "message", Type: "s", Direction: "in"}},
	}})

	got := result.String()
	if !strings.Contains(got, "Methods") || !strings.Contains(got, "Ping(message s)") {
		t.Fatalf("writeMethods() = %q, want method header and formatted args", got)
	}
}

func TestWriteProperties(t *testing.T) {
	var result strings.Builder
	writeProperties(&result, []introspect.Property{{
		Name:        "Status",
		Type:        "s",
		Access:      "read",
		Annotations: []introspect.Annotation{{Name: "org.example.Foo", Value: "bar"}},
	}})

	got := result.String()
	if !strings.Contains(got, "Properties") || !strings.Contains(got, "Name: Status") || !strings.Contains(got, "Type: s") || !strings.Contains(got, "Access: read") {
		t.Fatalf("writeProperties() = %q, want property details", got)
	}
}

func TestWriteSignals(t *testing.T) {
	var result strings.Builder
	writeSignals(&result, []introspect.Signal{{
		Name: "Changed",
		Args: []introspect.Arg{{Name: "message", Type: "s", Direction: "in"}},
	}})

	got := result.String()
	if !strings.Contains(got, "Signals") || !strings.Contains(got, "Changed(message s)") {
		t.Fatalf("writeSignals() = %q, want signal header and formatted args", got)
	}
}
