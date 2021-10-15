package licensemanager

// LicenseInfo represents the json returned from license server
type LicenseInfo struct {
	JavaClass  string    `json:"javaClass"`
	Restricted bool      `json:"restricted"`
	List       []License `json:"list"`
}
