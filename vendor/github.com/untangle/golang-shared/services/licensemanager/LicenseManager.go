package licensemanager

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/untangle/golang-shared/services/logger"
)

var config Config
var serviceStates []ServiceState

var errServiceNotFound error = errors.New("service_not_found")
var shutdownChannelLicense chan bool
var wg sync.WaitGroup
var watchDog *time.Timer

// Startup the license manager service.
// @param configOptions LicenseManagerConfig - a license manager config object used for configuring the service
func Startup(configOptions Config) {
	shutdownChannelLicense = make(chan bool)

	config = configOptions

	logger.Info("Starting the license service\n")

	serviceStates, err := loadServiceStates(config.ServiceStateLocation)
	if err != nil {
		logger.Warn("Unable to retrieve previous service state. %v\n", err)
	}

	if serviceStates == nil {
		// Gen a new state file
		serviceStates, err = saveServiceStates(config.ServiceStateLocation, config.ValidServiceHooks)
		if err != nil {
			logger.Warn("Unable to initialize service states file. %v\n", err)
			return
		}
	}

	// Set each service to its previous state.
	logger.Debug("States %+v\n", serviceStates)
	for _, o := range serviceStates {
		if _, err = findService(o.Name); err != nil {
			logger.Debug("Service %s not found. Err: %v", o.Name, err)
			continue
		}
		cmd := ServiceCommand{Name: o.Name}
		if o.IsEnabled {
			cmd.NewState = StateEnable
		} else {
			cmd.NewState = StateDisable
		}
		cmd.SetServiceState(true)
	}

	// restart licenses
	err = RefreshLicenses()
	if err != nil {
		logger.Warn("Not able to restart CLS: %v\n", err)
	}

	// watchdog for if CLS is alive
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchDog = time.NewTimer(config.WatchDogInterval)
		defer watchDog.Stop()
		for {
			select {
			case <-shutdownChannelLicense:
				logger.Info("Shutdown CLS watchdog\n")
				return
			case <-watchDog.C:
				// on watch dog seen, restart license server
				// shutdown license items if restart did not work
				logger.Warn("Watch seen\n")
				refreshErr := RefreshLicenses()
				if refreshErr != nil {
					logger.Warn("Couldn't restart CLS: %s\n", refreshErr)
					shutdownServices(config.LicenseLocation, config.ValidServiceHooks)
				} else {
					logger.Info("Restarted CLS from watchdog\n")
				}
				watchDog.Reset(config.WatchDogInterval)
			}
		}
	}()
}

// Shutdown is called when the service stops
func Shutdown() {
	logger.Info("Shutting down the license service\n")
	if shutdownChannelLicense != nil {
		close(shutdownChannelLicense)
		wg.Wait()
	}
	shutdownServices(config.LicenseLocation, config.ValidServiceHooks)
}

// GetLicenseDefaults gets the default validServiceStates
// @return []string - string array of service keys for CLS to use
func GetLicenseDefaults() []string {
	logger.Debug("GetLicenseDefaults()\n")
	keys := make([]string, len(config.ValidServiceHooks))
	i := 0
	for k := range config.ValidServiceHooks {
		keys[i] = k
		i++
	}
	watchDog.Reset(config.WatchDogInterval)
	return keys
}

// ClsIsAlive resets the watchdog interval for license <> service synchronization
func ClsIsAlive() {
	watchDog.Reset(config.WatchDogInterval)
}

// GetServiceStates gets the current Service States
// @return []ServiceState - array of current service states
func GetServiceStates() []ServiceState {
	return serviceStates
}

// RefreshLicenses restart the client licence service
func RefreshLicenses() error {
	output, err := exec.Command("/etc/init.d/clientlic", "restart").CombinedOutput()
	if err != nil {
		logger.Warn("license fetch failed: %s\n", err.Error())
		return err
	}
	if strings.Contains(string(output), "Command failed") {
		logger.Warn("license fetch failed: %s\n", string(output))
		err = errors.New(string(output))
		return err
	}
	return nil
}

// IsEnabled is called from API to see if service is currently enabled.
// @param serviceName string - the name of the service to check Enabled status of
func IsEnabled(serviceName string) (bool, error) {
	var serv ServiceHook
	var err error
	if serv, err = findService(serviceName); err != nil {
		return false, errServiceNotFound
	}
	return serv.Enabled(), nil
}

// GetLicenseDetails will use the current license location to load and return the license file
// @return LicenseInfo - the license info, containing license details
// @return error - associated errors
func GetLicenseDetails() (LicenseInfo, error) {

	var retLicense LicenseInfo

	// Load file
	jsonLicense, err := ioutil.ReadFile(config.LicenseLocation)

	// Unmarshal
	err = json.Unmarshal(jsonLicense, &retLicense)
	if err != nil {
		logger.Warn("Error unmarshalling licenseInfo: %s\n", err.Error())
		return retLicense, err
	}

	// Return
	return retLicense, nil
}

// shutdownServices iterates servicesToShutdown and calls the shutdown hook on them, and also removes the license file
// @param licenseFile string - the license file location
// @param servicesToShutdown map[string]ServiceHook - the services we want to shutdown
func shutdownServices(licenseFile string, servicesToShutdown map[string]ServiceHook) {
	err := os.Remove(licenseFile)
	err = ioutil.WriteFile(licenseFile, []byte("{\"list\": []}"), 0444)
	if err != nil {
		logger.Warn("Failure to write non-license file: %v\n", err)
	}
	for name := range servicesToShutdown {
		cmd := ServiceCommand{Name: name, NewState: StateDisable}
		cmd.SetServiceState(false)
	}
}

// findService is used to check if service is valid and return its hooks
// @param serviceName string - the name of the service
func findService(serviceName string) (ServiceHook, error) {
	service, ok := config.ValidServiceHooks[serviceName]
	if !ok {
		return ServiceHook{}, errServiceNotFound
	}
	return service, nil
}

// saveServiceStates stores the services in the service state file
// @param string fileLocation - the location of the service state file
// @param map[string]ServiceHook serviceHooks - map of service hooks to create ServiceStates for
// @return []ServiceState - the service state array
// @return error - associated errors
func saveServiceStates(fileLocation string, serviceHooks map[string]ServiceHook) ([]ServiceState, error) {
	var serviceStates = make([]ServiceState, 0)
	for name, o := range serviceHooks {
		serviceStates = append(serviceStates, ServiceState{Name: name, IsEnabled: o.Enabled()})
	}
	data, err := json.Marshal(serviceStates)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(fileLocation, data, 0644)
	if err != nil {
		return nil, err
	}
	return serviceStates, nil

}

// loadServiceStates retrieves the previously saved service state
// @param fileLocation - the location of the service states file
// @return []ServiceState - an array of service states, loaded from the file
// @return error - associated errors
func loadServiceStates(fileLocation string) ([]ServiceState, error) {
	var serviceStates = make([]ServiceState, 0)
	fileContent, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		logger.Warn("Not able to find service state file.\n", err)
		return nil, nil
	}

	err = json.Unmarshal(fileContent, &serviceStates)
	if err != nil {
		logger.Warn("Not able to read content of service state file.%v \n ", err)
		return nil, err
	}
	return serviceStates, nil
}
