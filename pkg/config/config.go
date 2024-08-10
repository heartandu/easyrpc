package config

// Config represents a common cross-application configuration.
type Config struct {
	Server Server
}

// Server represents a configuration of a remote server connection.
type Server struct {
	Address string
}
