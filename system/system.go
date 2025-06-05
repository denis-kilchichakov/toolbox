package system

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdownSignal() {
	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	signal := <-sigCh
	log.Printf("received %s", signal)
}
