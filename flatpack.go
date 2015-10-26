package flatpack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Getter represents a read-only repository of key/value pairs where the keys
// are ordered sequences of strings and the values are strings. It's analogous
// to a map[[]string]string but the data may be retrieved from a network or
// filesystem source, and the complex names may be treated as an indicator of
// hierarchy or containment within the data source, e.g. URL hierarchy on an
// HTTP k/v store, or directory hierarchy on a filesystem-based store.
type Getter interface {
	Get(name []string) (string, error)
}

// Unmarshal reads configuration data into a struct.
func Unmarshal(data Getter, dest interface{}) error {
	return unmarshal(data, []string{}, dest)
}

func unmarshal(data Getter, prefix []string, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		return fmt.Errorf("Cannot unmarshal: need pointer to struct, got %s", v.Kind().String())
	}

	vt := v.Type()

	if vt.Kind() != reflect.Struct {
		return fmt.Errorf("Invalid kind for %v: expected struct, got %s", prefix, vt.Kind().String())
	}

	// prepare field-name that we can reuse across fields
	name := make([]string, len(prefix)+1)
	copy(name, prefix)

	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)
		value := v.Field(i)

		name[len(name)-1] = field.Name
		err := read(data, name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func assign(dest reflect.Value, source string) (err error) {
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

// Unmarshal a single field by reading a string from the Getter, massaging it
// to the correct Type for that field, and assigning to the given Value.
func read(data Getter, name []string, value reflect.Value) error {
	kind := value.Type().Kind()

	var err error

	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.String:
		got, err := data.Get(name)
		if err == nil {
			err = assign(value, got)
		}
	case reflect.Array, reflect.Slice:
		got, err := data.Get(name)
		if err == nil {
			var raw []interface{}
			err = json.Unmarshal([]byte(got), &raw)
			if err == nil {
				value.Set(reflect.MakeSlice(value.Type(), len(raw), len(raw)))
				for i, elem := range raw {
					if err == nil {
						err = assign(value.Index(i), fmt.Sprintf("%v", elem))
					}
				}
			}
		}
	case reflect.Struct:
		unmarshal(data, name, value.Addr().Interface())
	default:
		// TODO something less impolite!!
		panic(fmt.Sprintf("Don't know how to deal with %v", value.Interface()))
	}

	return err
}
