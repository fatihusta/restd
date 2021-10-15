package licensemanager

// ServiceState is used to load and save AppState. Need to set each app to its previous state when starting up.
type ServiceState struct {
	Name      string `json:"name"`
	IsEnabled bool   `json:"enabled"`
}
