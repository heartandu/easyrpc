// Package config provides configuration management for the application.
//
//nolint:revive // Multiple public structs are necessary for configuration.
package config

// Config represents a common cross-application configuration.
type Config struct {
	Proto   Proto   `mapstructure:",squash" yaml:",inline"`
	Server  Server  `mapstructure:",squash" yaml:",inline"`
	TLS     TLS     `mapstructure:",squash" yaml:",inline"`
	Request Request `mapstructure:",squash" yaml:",inline"`
	Editor  Editor  `mapstructure:",squash" yaml:",inline"`
}

// Proto represents a set of proto files related configuration.
type Proto struct {
	ImportPaths []string `mapstructure:"import_paths" yaml:"import_paths" pflag:"import-path"`
	ImportAll   bool     `mapstructure:"import_all"   yaml:"import_all"   pflag:"import-all"`
	ProtoFiles  []string `mapstructure:"proto_files"  yaml:"proto_files"  pflag:"proto-file"`
}

// Server represents a configuration of a remote server connection.
type Server struct {
	Address    string `mapstructure:"address"    yaml:"address"    pflag:"address"`
	Reflection bool   `mapstructure:"reflection" yaml:"reflection" pflag:"reflection"`
	Web        bool   `mapstructure:"web"        yaml:"web"        pflag:"web"`
}

// TLS represents TLS configuration.
type TLS struct {
	Enabled bool   `mapstructure:"tls"    yaml:"tls"    pflag:"tls"`
	CACert  string `mapstructure:"cacert" yaml:"cacert" pflag:"cacert"`
	Cert    string `mapstructure:"cert"   yaml:"cert"   pflag:"cert"`
	Key     string `mapstructure:"key"    yaml:"key"    pflag:"key"`
}

// Request represents a request configuration.
type Request struct {
	Metadata map[string]string `mapstructure:"metadata" yaml:"metadata" pflag:"metadata,headersSlice"`
	Package  string            `mapstructure:"package"  yaml:"package"  pflag:"package"`
	Service  string            `mapstructure:"service"  yaml:"service"  pflag:"service"`
}

// Editor represents a message editor utility configuration.
type Editor struct {
	Cmd string `yaml:"-"`
}
