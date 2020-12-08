package main
import (
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
	"time"
	

	"github.com/untangle/restd/services/gind"
	"github.com/untangle/packetd/services/logger"
)

var shutdownFlag bool
func main() {
	userinfo, err := user.Current()
	if err != nil {
		panic(err)
	}

	userid, err := strconv.Atoi(userinfo.Uid)
	if err != nil {
		panic(err)
	}

	if userid != 0 {
		panic("This application must be run as root\n")
	}

	logger.Startup()

	setIsShutdown(false)
	gind.Startup()

	handleSignals()

	for !getShutdown() {
		select {
		case <-time.After(2 * time.Second):
			logger.Debug("Time")
		}
	}

	logger.Info("Shutdown restd logger\n")
	logger.Shutdown()

}

func setIsShutdown(flag bool) {
	shutdownFlag = flag
}

func getShutdown() bool {
	return shutdownFlag
}

func handleSignals() {
	// Add SIGINT & SIGTERM handler (exit)
	termch := make(chan os.Signal, 1)
	signal.Notify(termch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termch
		go func() {
			logger.Info("Received signal [%v]. Setting shutdown flag\n", sig)
			setIsShutdown(true)
		}()
	}()

	// Add SIGQUIT handler (dump thread stack trace)
	/*
	TODO
	quitch := make(chan os.Signal, 1)
	signal.Notify(quitch, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-quitch
			logger.Info("Received signal [%v]. Calling dumpStack()\n", sig)
			// TODO go dumpStack()
		}
	}()

	// Add SIGHUP handler (call handlers)
	hupch := make(chan os.Signal, 1)
	signal.Notify(hupch, syscall.SIGHUP)
	go func() {
		for {
			sig := <-hupch
			logger.Info("Received signal [%v]. Calling handlers\n", sig)
		}
	}()
	*/
}