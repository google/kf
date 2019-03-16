package kf

// ConfigErr is used to indicate that the returned error is due to a user's
// invalid configuration.
type ConfigErr struct {
	// Reason holds the error message.
	Reason string
}

// Error implements error.
func (e ConfigErr) Error() string {
	return e.Reason
}

// ConfigError returns true if the error is due to user error.
func ConfigError(err error) bool {
	_, ok := err.(ConfigErr)
	return ok
}
