package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hellobchain/memmq/cmd/memmq"
	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetLoggerWithoutName()

func main() {
	// handle client
	isCli := memmq.StartMain()
	// if cli mode exit
	if isCli {
		return
	}
	// listen for signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("MQ server stopped")
}
