package brian

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//信号
func hookSignals(app *Application) {
	sigChan := make(chan os.Signal)
	signal.Notify(
		sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)

	go func() {
		var sig os.Signal
		for {
			sig = <-sigChan
			switch sig.String() {
			case syscall.SIGQUIT.String():
				_ = app.GracefulStop(context.TODO())
			case syscall.SIGHUP.String(), syscall.SIGINT.String(), syscall.SIGTERM.String(), syscall.SIGKILL.String():
				_ = app.Stop() // terminate now
			}
			time.Sleep(time.Second * 3)
		}
	}()
}
