package flatpack

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// Getter represents a read-only repository of key/value pairs where the keys
// are ordered sequences of strings and the values are strings. It's analogous
// to a map[[]string]string, but the data may be retrieved from a network or
// filesystem source and the complex names may be treated as an indicator of
// hierarchy or containment within the data source, e.g. URL hierarchy on an
// HTTP k/v store, or directory hierarchy on a filesystem-based store.
type Getter interface {
	Get(name []string) (string, error)
}

// DataSource is the single source of configuration used for all calls to
// flatpack's singleton interface. When someone calls flatpack.Unmarshal(),
// the data comes from this source.
//
// By default, DataSource points to the process environment.
var DataSource Getter = &processEnvironment{os.LookupEnv}

// Unmarshal reads configuration data from the package's DataSource into
// a struct.
//
// This is the singleton/package-level interface to flatpack, for applications
// that want to use the default data source (process environment) or
// set a process-wide data source at startup.
//
// NOTE: flatpack currently lacks a non-singleton, object-level interface
// but it would be easy to add by simply making flatpack.new() into an exported
// func. I am deferring this commitment until there's a demand for it...
func Unmarshal(dest interface{}) error {
	return new(DataSource).Unmarshal(dest)
}

// Construct an Unmarshaller for the given data source.
// TODO export this some day when we have non-trivial data sources.
func new(source Getter) Unmarshaller {
	return &flatpack{source}
}

type flatpack struct {
	source Getter
}

// Unmarshal reads configuration data from some source into a struct.
func (f flatpack) Unmarshal(dest interface{}) error {
	return f.unmarshal([]string{}, dest)
}

// Read configuration source into a struct or sub-struct.
func (f flatpack) unmarshal(prefix []string, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		return fmt.Errorf("invalid dest parameter: need pointer to struct, got %s", v.Kind().String())
	}

	vt := v.Type()

	if vt.Kind() != reflect.Struct {
		return fmt.Errorf("invalid kind for %v: expected struct, got %s", prefix, vt.Kind().String())
	}

	// prepare field-name that we can reuse across fields
	name := make([]string, len(prefix)+1)
	copy(name, prefix)

	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)
		value := v.Field(i)

		name[len(name)-1] = field.Name
		err := f.read(name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// Coerce a value to a suitable Type and then assign it to a Value (either a
// struct field or an element of a slice).
func (f flatpack) assign(dest reflect.Value, source string) (err error) {
	kind := dest.Type().Kind()

	switch kind {
	case reflect.Bool:
		boolean, err := strconv.ParseBool(source)
		if err != nil {
			dest.SetBool(boolean)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, err := strconv.ParseInt(source, 10, int(dest.Type().Size()*8))
		if err == nil {
			dest.SetInt(number)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		number, err := strconv.ParseUint(source, 10, int(dest.Type().Size()*8))
		if err == nil {
			dest.SetUint(number)
		}
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(source, int(dest.Type().Size()*8))
		if err == nil {
			dest.SetFloat(number)
		}
	case reflect.String:
		if err == nil {
			dest.SetString(source)
		}
	default:
		err = fmt.Errorf("Cannot assign %v to a %s", source, dest.Type().Name())
	}

	return
}

// Set a single struct field by reading a string from the Getter, massaging it
// to the correct Type for that field, and assigning to the given Value.
func (f flatpack) read(name []string, value reflect.Value) error {
	kind := value.Type().Kind()

	var err error

	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.String:
		got, err := f.source.Get(name)
		if err == nil {
			err = f.assign(value, got)
		}
	case reflect.Array, reflect.Slice:
		got, err := f.source.Get(name)
		if err == nil {
			var raw []interface{}
			err = json.Unmarshal([]byte(got), &raw)
			if err == nil {
				value.Set(reflect.MakeSlice(value.Type(), len(raw), len(raw)))
				for i, elem := range raw {
					if err == nil {
						err = f.assign(value.Index(i), fmt.Sprintf("%v", elem))
					}
				}
			}
		}
	case reflect.Struct:
		f.unmarshal(name, value.Addr().Interface())
	default:
		err = fmt.Errorf("invalid value for %v;  unsupported type %v", name, value.Type().Name())
	}

	return err
}
