package exithandler

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Init accepts a callback function that will be invoked when program exits
// unexpectedly or is terminated by user. Used to perform cleanup: close DB
// connection, close network connections etc
func Init(cb func()) {
	sigs := make(chan os.Signal, 1)
	terminate := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Fatal("exit reason: ", sig)
		close(terminate)
	}()

	<-terminate
	cb()
}
