package core

import "fmt"

type (
	NotFoundError      string
	ThirdPartyError    string
	ErrLockTaken       string
	ConfigurationError string
)

func (e NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", string(e))
}

func (e ThirdPartyError) Error() string {
	return fmt.Sprintf("third-party error: %s", string(e))
}

func (e ErrLockTaken) Error() string {
	return fmt.Sprintf("lock taken: %s", string(e))
}

func (e ConfigurationError) Error() string {
	return fmt.Sprintf("missing or invalid config at path: %s", string(e))
}
