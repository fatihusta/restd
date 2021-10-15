package licensemanager

// ServiceHook struct is used to indicate start/stop/enabled hooks for services
type ServiceHook struct {
	Start   func()
	Stop    func()
	Enabled func() bool
}
