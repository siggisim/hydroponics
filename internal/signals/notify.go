package signals

import (
	"os"
	"os/signal"
)

func Notify(ch chan os.Signal) {
	signal.Notify(ch, ExitSignals...)
}
