# EasyRPC

EasyRPC is an easy-to-use gRPC client.

The main purpose of this CLI utility is to offer a user-friendly interface with completions and support for gRPC-Web
for manual inspection of gRPC APIs.
EasyRPC is influenced by the utilities [`grpcurl`](https://github.com/fullstorydev/grpcurl) and
[`evans`](https://github.com/ktr0731/evans), and aims to combine the two different approaches (basic CLI and REPL) into
a more convenient tool for users.

[TOC]

## Installation

### From source

To be able to install from source, you will need to install the [Go SDK](https://go.dev/dl/).
Go v1.23 or later is required.
After installation, run the following:

```shell
go install github.com/heartandu/easyrpc/cmd/easyrpc@latest
```

Ensure that your `GOBIN` directory (typically located at `$HOME/go/bin`) is added to the `PATH`,
or verify that the installed binary is accessible in one of the standard `PATH` locations on your system.

### Autocompletion

To begin using it, you must register the autocompletions script for your preferred shell.
Currently, the following shells are supported: `bash`, `fish`, `zsh`, and `powershell`.
Please refer to the `easyrpc completion -h` command help to learn how to register completions for specific shells.

## Usage

To view the full list of available commands and flags, run `easyrpc -h`.

### Invoking RPCs

Make a simple unary call:

```shell
# No TLS, empty message, using local proto files
$ easyrpc c -a localhost:12345 -i path/to/proto -p example.proto example.package.Service.Method
{
  "msg": ""
}

# Using TLS, with message, using server reflection
$ easyrpc c -a localhost:12345 -r --tls -d '{"msg":"hello"}' example.package.Service.Method
{
  "msg": "hello"
}

# Multiple protobuf import paths and files
$ easyrpc c -a localhost:12345 -i path/to/proto -i other/path/to/proto -p server/v1/foo.proto -p client/v2/bar.proto example.package.Service.Method
{
  "msg": ""
}
```

### Streaming RPCs

Making streaming calls.

```shell
# Client streaming
$ easyrpc c -a localhost:12345 -r example.package.Service.ClientStreaming -d '{"msg":"1"}{"msg":"2"}{"msg":"3"}'
{
  "msgs": [
    "1",
    "2",
    "3"
  ]
}

# Server streaming
$ easyrpc c -a localhost:12345 -r example.package.Service.ServerStreaming -d '{"msgs":["1","2","3"]}'
{
  "msg": "1"
}
{
  "msg": "2"
}
{
  "msg": "3"
}

# Bidirectional streaming
$ easyrpc c -a localhost:12345 -r example.package.Service.BidiStreaming -d '{"msg":"1"}{"msg":"2"}{"msg":"3"}'
{
  "msg": "1"
}
{
  "msg": "2"
}
{
  "msg": "3"
}
```

### TLS

EasyRPC supports TLS termination, including mutual TLS.

```shell
# TLS call
$ easyrpc c -a localhost:12345 -r --tls example.package.Service.Method

# Mutual TLS termination
$ easyrpc c -a localhost:12345 -r --cert path/to/localhost.crt --key path/to/localhost.key --tls example.package.Service.Method

# Using custom root certificate
$ easyrpc c -a localhost:12345 -r --cacert path/to/root.crt --tls example.package.Service.Method
```

### Metadata

You can provide metadata to send with the request.

```shell
# Single header
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -H 'Authorization=Bearer token'

# Multiple headers
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -H 'Authorization=Bearer token' -H 'X-Real-Ip=0.0.0.0'
```

### Input data

There are also multiple ways of providing request message data.

```shell
# Providing data in the flag itself
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -d '{"msg":"hello"}'

# Receiving the data from stdin
$ echo '{"msg":"hello"}' | easyrpc -a localhost:12345 -r example.package.Service.Method -d -

# Reading the data from file
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -d @~/some/path/request.json
```

### Autocompletion

You can use autocompletion to fill in the method name.

```shell
# Inputting this
$ easyrpc c -a localhost:12345 -r Me[tab]

# Will result in
$ easyrpc c -a localhost:12345 -r example.package.Service.Method

# Or inputing this
$ easyrpc c -i path/to/proto -p example.proto Me[tab]

# Will result in
$ easyrpc c -i path/to/proto -p example.proto example.package.Service.Method
```

Autocompletion works with both local proto files and server reflection.
However, it is necessary to provide one of the protobuf sources in order for completions to work.

You can also set the `--package` and `--service` names to reduce the amount of text to input when requesting different
methods.
Autocompletion will also consider these flags.
For example:

```shell
# We mostly work with the "example.package" package.
$ easyrpc c -a localhost:12345 -r --package example.package Me[tab]

# The input above will result in
$ easyrpc c -a localhost:12345 -r --package example.package Service.Method

# Or we can also provide specific "service"
$ easyrpc c -a localhost:12345 -r --package example.package --service Service Me[tab]

# And that will result in
$ easyrpc c -a localhost:12345 -r --package example.package --service Service Method
```

Note that the `package` and `service` flags can also be autocompleted if one of the protobuf sources is provided.

### Configuration files

In order to reduce the amount of terminal boilerplate, you can store commonly used parameters in a configuration file.
The default locations for configuration files are `$HOME/.easyrpc.yaml` and `.easyrpc.yaml` in the working directory.
You can also specify the configuration file explicitly using the `--config` flag.

For example, given the configuration file:

```yaml
address: localhost:12345
import_paths:
    - ~/path/to/proto
proto_files:
    - example.proto
package: example.package
service: Service
metadata:
    authorization: Bearer token
```

The actual command will look something like this:

```shell
$ easyrpc c Method -d '{"msg":"hello"}'
```

Autocompletion also works with configuration files.

```shell
# Inputting this
$ easyrpc c -d '{"msg":"hello"}' Me[tab]

# Will result in
$ easyrpc c -d '{"msg":"hello"}' Method
```

All configuration files and CLI flags are loaded and merged simultaneously.
The precedence of the locations is as follows:

- CLI flags
- Configuration file from `--config` flag
- `./.easyrpc.yaml`
- `$HOME/.easyrpc.yaml`

You can initialize the configuration with empty values in the current working directory by running `easyrpc config init`.
If you want to inspect the resulting configuration that will be used by `easyrpc`, run `easyrpc config dump`.

### gRPC-Web

EasyRPC supports a gRPC-Web translation layer for both unary and streaming calls.
Unary calls are made as HTTP 1.1 requests, while streaming calls are implemented using websockets.
The gRPC-Web implementation is compatible with the [improbable-eng/grpc-web](https://github.com/improbable-eng/grpc-web)
and [envoy proxy](https://www.envoyproxy.io/) implementations.
EasyRPC also supports TLS as well as mutual TLS termination over the gRPC-Web translation layer.

To enable the translation layer, use `--web` or `-w` flag.
For example:

```shell
# Plain text call
$ easyrpc c -a localhost:12345 -r -w example.package.Service.Method

# TLS call
$ easyrpc c -a localhost:12345 -r -w --tls example.package.Service.Method

# Call to a prefixed endpoint
$ easyrpc c -a localhost:12345/grpc-web -r -w example.package.Service.Method
```
