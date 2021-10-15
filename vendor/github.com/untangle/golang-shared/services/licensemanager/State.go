package licensemanager

import (
	"errors"
	"strings"
)

// State is an enum used for indicating service state
type State int

// Types of commands that we accept.
const (
	StateEnable State = iota
	StateDisable
)

// FromString is used to convert the string command into the current State
// @param cmd - the command string
// @return error - associated errors
func (state *State) FromString(cmd string) error {

	switch lowerCmd := strings.ToLower(cmd); lowerCmd {
	case "enable":
		*state = StateEnable
	case "disable":
		*state = StateDisable
	default:
		return errors.New("Unable to parse state")
	}

	return nil
}
