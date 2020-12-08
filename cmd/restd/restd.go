package main
import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"
	"time"
	

	"github.com/untangle/restd/services/gind"
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
		panic("This application must be run as root")
	}

	setIsShutdown(false)
	gind.Startup()

	handleSignals()

	for !getShutdown() {
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("Time")
		}
	}

}

func setIsShutdown(flag bool) {
	shutdownFlag = flag
}

func getShutdown() bool {
	return shutdownFlag
}

func handleSignals() {
	// TODO fmt-->logger.Info
	// Add SIGINT & SIGTERM handler (exit)
	termch := make(chan os.Signal, 1)
	signal.Notify(termch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termch
		go func() {
			fmt.Println("Received signal [%v]. Setting shutdown flag\n", sig)
			setIsShutdown(true)
		}()
	}()

	// Add SIGQUIT handler (dump thread stack trace)
	quitch := make(chan os.Signal, 1)
	signal.Notify(quitch, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-quitch
			fmt.Println("Received signal [%v]. Calling dumpStack()\n", sig)
			// TODO go dumpStack()
		}
	}()

	// Add SIGHUP handler (call handlers)
	hupch := make(chan os.Signal, 1)
	signal.Notify(hupch, syscall.SIGHUP)
	go func() {
		for {
			sig := <-hupch
			fmt.Println("Received signal [%v]. Calling handlers\n", sig)
		}
	}()
}