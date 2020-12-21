package main
import (
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	

	"github.com/jsommerville-untangle/golang-shared/services/logger"
	"github.com/untangle/restd/services/gind"
	"github.com/untangle/restd/services/messenger"
)

var shutdownFlag uint32
var shutdownChannel = make (chan bool)

func main() {
	// Check we are root user
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

	// Start up logger
	logger.Startup()
	logger.Info("Starting up restd...\n")

	// Start services
	startServices()

	handleSignals()

	for !GetShutdownFlag() {
		select {
		case <-shutdownChannel:
			logger.Info("Shutdown channel initiated... %v\n", GetShutdownFlag())
		case <-time.After(2 * time.Second):
			logger.Info("restd is running...\n")
			// logger.Info("\n")
			// printStats()
		}
	}

	logger.Info("Shutdown restd...\n")
	stopServices()

}

func startServices() {
	gind.Startup()
	messenger.Startup()
}

func stopServices() {
	gind.Shutdown()
	messenger.Shutdown()
	logger.Shutdown()
}

func handleSignals() {
	// Add SIGINT & SIGTERM handler (exit)
	termch := make(chan os.Signal, 1)
	signal.Notify(termch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-termch
		go func() {
			logger.Info("Received signal [%v]. Setting shutdown flag\n", sig)
			SetShutdownFlag()
		}()
	}()

	// Add SIGQUIT handler (dump thread stack trace)
	quitch := make(chan os.Signal, 1)
	signal.Notify(quitch, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-quitch
			logger.Info("Received signal [%v]. Calling dumpStack()\n", sig)
			go dumpStack()
		}
	}()
}

// dumpStack to /tmp/restd.stack and log
func dumpStack() {
	buf := make([]byte, 1<<20)
	stacklen := runtime.Stack(buf, true)
	ioutil.WriteFile("/tmp/restd.stack", buf[:stacklen], 0644)
	logger.Warn("Printing Thread Dump...\n")
	logger.Warn("\n\n%s\n\n", buf[:stacklen])
	logger.Warn("Thread dump complete.\n")
}

// prints some basic stats about packetd
func printStats() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	logger.Info("Memory Stats:\n")
	logger.Info("Memory Alloc: %d kB\n", (mem.Alloc / 1024))
	logger.Info("Memory TotalAlloc: %d kB\n", (mem.TotalAlloc / 1024))
	logger.Info("Memory HeapAlloc: %d kB\n", (mem.HeapAlloc / 1024))
	logger.Info("Memory HeapSys: %d kB\n", (mem.HeapSys / 1024))
}

// GetShutdownFlag returns the shutdown flag for kernel
func GetShutdownFlag() bool {
	if atomic.LoadUint32(&shutdownFlag) != 0 {
		return true
	}
	return false
}

// SetShutdownFlag sets the shutdown flag for kernel
func SetShutdownFlag() {
	shutdownChannel <- true
	atomic.StoreUint32(&shutdownFlag, 1)
}