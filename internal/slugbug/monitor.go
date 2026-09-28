package slugbug

import "fmt"

const Monitor Operation = "Monitor signal(s)"

func (s *Slugbug) MonitorSignal(service string, signal string) {
	fmt.Printf("Monitoring signal: '%s' on service: '%s'\n", signal, service)
}
