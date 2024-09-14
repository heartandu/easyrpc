package config

// Config represents a common cross-application configuration.
type Config struct {
	Proto   Proto   `mapstructure:",squash"`
	Server  Server  `mapstructure:",squash"`
	TLS     TLS     `mapstructure:",squash"`
	Request Request `mapstructure:",squash"`
}

// Proto represents a set of proto files related configuration.
type Proto struct {
	ImportPaths []string `mapstructure:"import_paths"`
	ProtoFiles  []string `mapstructure:"proto_files"`
}

// Server represents a configuration of a remote server connection.
type Server struct {
	Address    string `mapstructure:"address"`
	Reflection bool   `mapstructure:"reflection"`
	Web        bool   `mapstructure:"web"`
}

type TLS struct {
	Enabled bool   `mapstructure:"tls"`
	CACert  string `mapstructure:"cacert"`
	Cert    string `mapstructure:"cert"`
	Key     string `mapstructure:"key"`
}

// Request represents a request configuration.
type Request struct {
	Metadata map[string]string `mapstructure:"metadata"`
	Package  string            `mapstructure:"package"`
	Service  string            `mapstructure:"service"`
}
