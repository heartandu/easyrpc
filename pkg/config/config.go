package config

// Config represents a common cross-application configuration.
type Config struct {
	Server Server
	Proto  Proto
}

// Proto represents a set of proto files related configuration.
type Proto struct {
	ImportPaths []string `mapstructure:"import_paths"`
	ProtoFiles  []string `mapstructure:"proto_files"`
}

// Server represents a configuration of a remote server connection.
type Server struct {
	Address    string
	Reflection bool
}
