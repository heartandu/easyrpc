package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var (
	ErrNotSliceValue        = errors.New("value must be a SliceValue")
	ErrInvalidHeaderFormat  = errors.New(`value must be in "key: value" format`)
	ErrUnknownDecodeFunc    = errors.New("unknown decode function")
	ErrUnsupportedSliceElem = errors.New("unsupported slice element type")
	ErrUnsupportedFieldType = errors.New("unsupported field type")
	ErrGetBoolFailed        = errors.New("failed to get bool flag value")
	ErrGetStringSliceFailed = errors.New("failed to get string slice flag value")
	ErrGetStringFailed      = errors.New("failed to get string flag value")
	ErrInvalidYAML          = errors.New("failed to decode config file")
)

// Decoder decodes configuration from files and command-line flags.
type Decoder struct {
	fs    afero.Fs
	cmd   *cobra.Command
	files []string
}

// NewDecoder creates a new Decoder instance.
func NewDecoder(fs afero.Fs, cmd *cobra.Command, configFileNames []string) *Decoder {
	return &Decoder{
		fs:    fs,
		cmd:   cmd,
		files: configFileNames,
	}
}

// Decode reads configuration from files and flags, returning a populated Config.
func (d *Decoder) Decode() (cfg Config, err error) {
	result := make(map[string]any)

	for _, file := range d.files {
		fileMap, err := d.readFile(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return Config{}, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		mergeMaps(result, fileMap)
	}

	if d.cmd != nil {
		flagMap, err := d.readFlags(d.cmd.PersistentFlags())
		if err != nil {
			return Config{}, fmt.Errorf("failed to process persistent flags: %w", err)
		}

		mergeMaps(result, flagMap)
	}

	if err := mapstructure.Decode(result, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %w", err)
	}

	if editor, ok := os.LookupEnv("EDITOR"); ok {
		cfg.Editor.Cmd = editor
	}

	return cfg, nil
}

func (d *Decoder) readFile(name string) (map[string]any, error) {
	f, err := d.fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", name, err)
	}
	defer f.Close()

	fileMap := make(map[string]any)
	if err := yaml.NewDecoder(f).Decode(&fileMap); err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrInvalidYAML, name, err)
	}

	return fileMap, nil
}

func mergeMaps(dst, src map[string]any) {
	for key, value := range src {
		if srcMap, ok := value.(map[string]any); ok {
			if dstMap, ok := dst[key].(map[string]any); ok {
				mergeMaps(dstMap, srcMap)
			} else {
				dst[key] = srcMap
			}
		} else {
			dst[key] = value
		}
	}
}

func (d *Decoder) readFlags(pflags *pflag.FlagSet) (map[string]any, error) {
	result := make(map[string]any)
	cfgType := reflect.TypeFor[Config]()

	if err := d.processCfgType(pflags, cfgType, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *Decoder) processCfgType(pflags *pflag.FlagSet, t reflect.Type, result map[string]any) (err error) {
	for i := range t.NumField() {
		field := t.Field(i)

		if field.Type.Kind() == reflect.Struct {
			//nolint:govet // err shadowing is accepted.
			if err := d.processCfgType(pflags, field.Type, result); err != nil {
				return err
			}

			continue
		}

		flagName, decodeFunc := parsePflagTag(&field)
		if flagName == "" {
			continue
		}

		resultKey := mapstructureKey(&field)
		if resultKey == "" {
			continue
		}

		flag := pflags.Lookup(flagName)
		if flag == nil || !flag.Changed {
			continue
		}

		result[resultKey], err = decodeValue(pflags, field.Type, flagName, decodeFunc)
		if err != nil {
			return fmt.Errorf("%s: %w", resultKey, err)
		}
	}

	return nil
}

func parsePflagTag(field *reflect.StructField) (flagName, decodeFunc string) {
	pflagTag := field.Tag.Get("pflag")
	if pflagTag == "" {
		return "", ""
	}

	tagParts := strings.SplitN(pflagTag, ",", 2)
	flagName = tagParts[0]

	if len(tagParts) > 1 {
		decodeFunc = tagParts[1]
	}

	return flagName, decodeFunc
}

func mapstructureKey(field *reflect.StructField) string {
	mapstructureParts := strings.SplitN(field.Tag.Get("mapstructure"), ",", 2)
	if len(mapstructureParts) == 0 {
		return ""
	}

	return mapstructureParts[0]
}

func decodeValue(pflags *pflag.FlagSet, fieldType reflect.Type, flagName, funcName string) (any, error) {
	if funcName != "" {
		return decodeWithFunc(pflags, flagName, funcName)
	}

	return decodeByType(pflags, fieldType, flagName)
}

func decodeWithFunc(pflags *pflag.FlagSet, flagName, funcName string) (any, error) {
	if funcName == "headersSlice" {
		return headersSlice(pflags.Lookup(flagName).Value)
	}

	return nil, fmt.Errorf("%w: %s", ErrUnknownDecodeFunc, funcName)
}

func headersSlice(value pflag.Value) (map[string]any, error) {
	sliceValue, ok := value.(pflag.SliceValue)
	if !ok {
		return nil, ErrNotSliceValue
	}

	out := make(map[string]any)

	for _, pair := range sliceValue.GetSlice() {
		kv := strings.SplitN(pair, ": ", 2)
		if len(kv) != 2 {
			return nil, ErrInvalidHeaderFormat
		}

		out[kv[0]] = kv[1]
	}

	return out, nil
}

func decodeByType(pflags *pflag.FlagSet, fieldType reflect.Type, flagName string) (any, error) {
	switch fieldType.Kind() {
	case reflect.Bool:
		return getBoolFlag(pflags, flagName)
	case reflect.String:
		return getStringFlag(pflags, flagName)
	case reflect.Slice:
		return getStringSliceFlag(pflags, fieldType, flagName)
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedFieldType, fieldType)
	}
}

func getBoolFlag(pflags *pflag.FlagSet, flagName string) (bool, error) {
	val, err := pflags.GetBool(flagName)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrGetBoolFailed, err)
	}

	return val, nil
}

func getStringFlag(pflags *pflag.FlagSet, flagName string) (string, error) {
	val, err := pflags.GetString(flagName)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGetStringFailed, err)
	}

	return val, nil
}

func getStringSliceFlag(pflags *pflag.FlagSet, fieldType reflect.Type, flagName string) (any, error) {
	if fieldType.Elem().Kind() != reflect.String {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedSliceElem, fieldType.Elem())
	}

	val, err := pflags.GetStringSlice(flagName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGetStringSliceFailed, err)
	}

	return val, nil
}
