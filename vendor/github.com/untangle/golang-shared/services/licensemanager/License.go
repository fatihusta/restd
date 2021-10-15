package licensemanager

// License is the struct representing each individual license
type License struct {
	UID         string `json:"UID"`
	Type        string `json:"type"`
	End         int64  `json:"end"`
	Start       int64  `json:"start"`
	Seats       int64  `json:"seats" default:"-1"`
	DisplayName string `json:"displayName"`
	Key         string `json:"key"`
	KeyVersion  int    `json:"keyVersion"`
	Name        string `json:"name"`
	JavaClass   string `json:"javaClass"`
	Valid       bool   `json:"valid" default:"false"`
}
