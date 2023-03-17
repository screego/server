package mode

const (
	// Dev for development mode.
	Dev = "dev"
	// Prod for production mode.
	Prod = "prod"
)

var mode = Dev

// Set sets the new mode.
func Set(newMode string) {
	mode = newMode
}

// Get returns the current mode.
func Get() string {
	return mode
}
