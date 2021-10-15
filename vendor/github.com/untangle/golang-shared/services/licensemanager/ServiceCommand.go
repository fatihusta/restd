package licensemanager

import "github.com/untangle/golang-shared/services/logger"

// ServiceCommand is used for setting the service state
type ServiceCommand struct {
	Name     string `json:"name"`
	NewState State  `json:"command"`
}

// SetServiceState sets the desired state of an service
// @param save bool - if we should store the service state or not
// @return error - associated errors
func (cmd *ServiceCommand) SetServiceState(save bool) error {
	var err error
	var service ServiceHook
	logger.Debug("Setting state for service %s to %v\n", cmd.Name, cmd.NewState)
	if service, err = findService(cmd.Name); err != nil {
		return errServiceNotFound
	}

	switch cmd.NewState {
	case StateEnable:
		service.Start()
	case StateDisable:
		service.Stop()
	}
	if save {
		serviceStates, err = saveServiceStates(config.ServiceStateLocation, config.ValidServiceHooks)
	}
	return err
}
