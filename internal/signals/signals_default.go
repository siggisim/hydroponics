// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd

package signals

import (
	"os"
)

var ExitSignals = []os.Signal{
	os.Interrupt,
}
