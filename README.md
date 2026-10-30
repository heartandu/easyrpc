# EasyRPC

EasyRPC is an easy-to-use gRPC client.

The main purpose of this CLI utility is to offer a user-friendly interface with completions and support for gRPC-Web
for manual inspection of gRPC APIs.
EasyRPC is influenced by the utilities [`grpcurl`](https://github.com/fullstorydev/grpcurl) and
[`evans`](https://github.com/ktr0731/evans), and aims to combine the two different approaches (basic CLI and REPL) into
a more convenient tool for users.


<!-- mtoc-start -->

* [Installation](#installation)
  * [Homebrew](#homebrew)
  * [Binaries](#binaries)
  * [Source](#source)
  * [Register autocompletion](#register-autocompletion)
* [Usage](#usage)
  * [Invoking RPCs](#invoking-rpcs)
  * [Streaming RPCs](#streaming-rpcs)
  * [TLS](#tls)
  * [Metadata](#metadata)
  * [Input data](#input-data)
  * [Request data preparation](#request-data-preparation)
  * [Edit request before call/printout](#edit-request-before-callprintout)
  * [Listing available RPCs](#listing-available-rpcs)
  * [Autocompletion](#autocompletion)
  * [Configuration files](#configuration-files)
  * [gRPC-Web](#grpc-web)

<!-- mtoc-end -->

## Installation

### Homebrew

On macOS and Linux, `easyrpc` is available via [Homebrew](https://brew.sh/) Cask:

```shell
brew install --cask heartandu/easyrpc/easyrpc
```

### Binaries

Download the preferred binary from the [releases](https://github.com/heartandu/easyrpc/releases) page.

### Source

To be able to install from source, you will need to install the [Go SDK](https://go.dev/dl/).
Go v1.24.6 or later is required.
After installation, run the following:

```shell
go install github.com/heartandu/easyrpc/cmd/easyrpc@latest
```

Ensure that your `GOBIN` directory (typically located at `$HOME/go/bin`) is added to the `PATH`,
or verify that the installed binary is accessible in one of the standard `PATH` locations on your system.

### Register autocompletion

To begin using it, you must register the autocompletion script for your preferred shell.
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

# Load all proto files from all of the import paths (proto files must have a ".proto" extension)
$ easyrpc c -a localhost:12345 -i path/to/proto -i other/path/to/proto --import-all example.package.Service.Method
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
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -H 'Authorization: Bearer token'

# Multiple headers
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -H 'Authorization: Bearer token' -H 'X-Real-Ip: 0.0.0.0'
```

### Input data

There are also multiple ways of providing request message data.

```shell
# Providing data in the flag itself
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -d '{"msg":"hello"}'

# Receiving the data from stdin
$ echo '{"msg":"hello"}' | easyrpc c -a localhost:12345 -r example.package.Service.Method -d -

# Reading the data from file
$ easyrpc c -a localhost:12345 -r example.package.Service.Method -f ~/some/path/request.json
```

> [!NOTE]
> Only one of `-d` or `-f` flags may be set at the same time. They are mutually exclusive.

### Request data preparation

You can prepare a selected method request for future reuse. The request command must be supplied with at least one
source of protobuf descriptors: either protobuf files or remote server with reflection enabled. If both are provided,
reflection descriptors take precedence.

```shell
# Request command takes a request message from a specified method, and prints unpopulated fields (only on the top level)
# to standard output
$ easyrpc r -i path/to/proto -p example.proto example.package.Service.Method
{
  "id": 123,
  "msg": "some message",
  "nestedMessage": null
}

# Works with reflect, connection setup is the same as in call command (supports TLS and gRPC-Web)
$ easyrpc r -a localhost:12345 -r example.package.Service.Method

# Save output to a file (-o flag)
$ easyrpc r -i path/to/proto -p example.proto -o request.json example.package.Service.Method
```

### Edit request before call/printout

You can edit the request data before call or request printout. The editor can be set in the `EDITOR` environment
variable. If no editor has been set, the system default will be used instead.

> [!NOTE]
> If you're using a GUI editor, make sure to use `--wait` or `-w` flag if the editor supports it. This way the command
> will wait until the temporary file is closed before proceeding. Otherwise, EasyRPC will continue operation immediately
> and no changes to the temporary file will be respected.

```shell
# Add -e (--edit) flag to open a temporary request file in an external editor. Requests edited contain $schema file link
# with a temporary JSON Schema for selected method request message which allows JSON LSP to provide field autocompletion
# and value validation.
$ easyrpc c example.package.Service.Method -e

# Pass an existing request and then edit that request before call.
$ easyrpc c example.package.Service.Method -f request.json -e

# Edit also works with request command which helps with request preparation for future reuse.
$ easyrpc r example.package.Service.Method -f request.json -e -o modified_request.json

# If you supply more than one message for a client streaming method, each message will be opened for editing in
# sequence.
$ easyrpc c example.package.Service.Method -d '{"msg":"test1"}{"msg":"test2"}{"msg":"test3"}' -e

# It is possible to open an editor with pipes.
$ easyrpc r example.package.Service.Method -e | jq
```

### Listing available RPCs

Before invoking an RPC, you can list the methods available from the configured protobuf source.
The command uses the same descriptor source options as `call` and `request`: local proto files,
`--import-all`, server reflection, or the corresponding settings in a [configuration file](#configuration-files).

```shell
# List methods using server reflection
$ easyrpc ls -a localhost:12345 -r
example.package.Service.Method
example.package.Service.OtherMethod
```

You can also use the `--package` and `--service` flags to narrow down the output the same way as
with [autocompletion](#autocompletion). Matching parts are omitted from the printed names.

```shell
# Limit output to a specific service
$ easyrpc list -a localhost:12345 -r --package example.package --service Service
Method
OtherMethod
```

The `list` output is handy for integration with external fuzzy finders. For example, with `fzf`
in Zsh you can define a custom completion widget for `easyrpc call` and `easyrpc request`:

```shell
_fzf_complete_easyrpc() {
    local -a tokens
    tokens=(${(z)1})
    case "${tokens[-1]}" in
        call|c|request|r)
            _fzf_complete --reverse --no-preview -- "$@" < <(easyrpc ls) ;;
        *)
            _fzf_path_completion "$prefix" "$1" ;;
    esac
}
```

### Autocompletion

You can use autocompletion to fill in the method name.

```shell
# Inputting this
$ easyrpc c -a localhost:12345 -r Me[tab]

# Will result in
$ easyrpc c -a localhost:12345 -r example.package.Service.Method

# Or inputting this
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

Most flags are replaced by values from higher precedence locations, except for metadata.
Metadata values are merged individually by key, allowing you to set common values in a configuration file
while overriding dynamic ones from flags.

You can initialize the configuration with empty values in the current working directory by running `easyrpc config init`.
If you want to inspect the resulting configuration that will be used by `easyrpc`, run `easyrpc config dump`.

For configuration autocompletion and validation you can use the
[easyrpc.schema.json](https://raw.githubusercontent.com/heartandu/easyrpc/refs/heads/master/easyrpc.schema.json)
JSON Schema with your preferred editor or LSP.

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
