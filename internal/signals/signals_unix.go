// +build darwin dragonfly freebsd linux netbsd openbsd

package signals

import (
	"os"
	"syscall"
)

var ExitSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}
