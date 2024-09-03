package config

// Config represents a common cross-application configuration.
type Config struct {
	Request Request `mapstructure:"request"`
	Server  Server  `mapstructure:"server"`
	Proto   Proto   `mapstructure:"proto"`
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
	TLS        bool   `mapstructure:"tls"`
}

type Request struct {
	CACert  string `mapstructure:"cacert"`
	Cert    string `mapstructure:"cert"`
	CertKey string `mapstructure:"cert_key"`
}
