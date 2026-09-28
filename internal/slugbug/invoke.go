package slugbug

import "fmt"

const Invoke Operation = "Invoke a method"

func (s *Slugbug) InvokeMethod(service string, method string) error {
	fmt.Printf("Invoking method '%s' on service '%s'\n", method, service)
	return nil
}
