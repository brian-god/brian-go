package brian

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// +build windows
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
			switch sig {
			case syscall.SIGQUIT:
				_ = app.GracefulStop(context.TODO())
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL:
				_ = app.Stop() // terminate now
			}
			time.Sleep(time.Second * 3)
		}
	}()
}
