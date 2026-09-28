package slugbug

import (
	"fmt"
	"slices"
	"strings"

	"github.com/godbus/dbus/v5/introspect"
)

const Inspect Operation = "Inspect a service"

// some services are commonly returned during introspection
// but we don't want to offer them to the user every time
var SkipInspectionServices = []string{
	"org.freedesktop.DBus.Introspectable",
	"org.freedesktop.DBus.Properties",
	"org.freedesktop.DBus.Peer",
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
