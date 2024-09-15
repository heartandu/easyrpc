package config

// Config represents a common cross-application configuration.
type Config struct {
	Proto   proto   `mapstructure:",squash"`
	Server  server  `mapstructure:",squash"`
	TLS     tls     `mapstructure:",squash"`
	Request request `mapstructure:",squash"`
	Editor  editor  `mapstructure:",squash"`
}

// proto represents a set of proto files related configuration.
type proto struct {
	ImportPaths []string `mapstructure:"import_paths"`
	ProtoFiles  []string `mapstructure:"proto_files"`
}

// server represents a configuration of a remote server connection.
type server struct {
	Address    string `mapstructure:"address"`
	Reflection bool   `mapstructure:"reflection"`
	Web        bool   `mapstructure:"web"`
}

type tls struct {
	Enabled bool   `mapstructure:"tls"`
	CACert  string `mapstructure:"cacert"`
	Cert    string `mapstructure:"cert"`
	Key     string `mapstructure:"key"`
}

// request represents a request configuration.
type request struct {
	Metadata map[string]string `mapstructure:"metadata"`
	Package  string            `mapstructure:"package"`
	Service  string            `mapstructure:"service"`
}

// editor represents a message editor utility configuration.
type editor struct {
	Cmd string `mapstructure:"editor"`
}
