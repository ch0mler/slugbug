package slugbug

import (
	"testing"
)

func TestCloseConnectionWithoutConnection(t *testing.T) {
	sb := NewSlugBug(false, false, false, false)
	sb.CloseConnection()
}

// func TestCloseConnection(t *tesing.T) {

// }
