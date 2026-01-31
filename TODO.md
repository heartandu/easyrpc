# "Schema" call patterns

## Default

```shell
easyrpc schema package.Message
```

Writes a jsonschema file next to a protobuf file with the message.

## Output (-o) flag

```shell
easyrpc schema package.Message -o filepath
```

Same as before, but put the jsonschema into `filepath`.

## Edit flag (-e)

```shell
easyrpc schema package.Message -o filepath -e
```

Same as before, but open $EDITOR with the following json:

```json
{
  "$schema": "filepath"
}
```

The editor must not have a predetermined file name, so the user will be able to save their file in the workdir (or
anywhere else).

# "Call" call pattern

```shell
easyrpc c SomeService.SomeMethod -e
```

Opens $EDITOR with jsonschema in a temporary file. The file must be deleted after the call.
