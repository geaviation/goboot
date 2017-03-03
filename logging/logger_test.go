package logging

import (
	"testing"
)

func TestLog(t *testing.T) {
	//t.Skipf("skipping...")

	//Logrus has six logging levels: Debug, Info, Warning, Error, Fatal and Panic.
	var log = Logger()

	log.Debug("Useful debugging information.")
	log.Info("Something noteworthy happened!")
	log.Warn("You should probably take a look at this.")
	log.Error("Something failed but I'm not quitting.")
	// Calls panic() after logging
	//log.Panic("I'm bailing.")
	// Calls os.Exit(1) after logging
	//log.Fatal("Bye.")
}
