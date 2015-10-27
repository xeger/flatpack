package flatpack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// unmarshaller represents an object that is capable of unmarshalling
// configuration data into destination structures. It encapsulates the
// source of the data as well as any options pertaining to data ccess.
//
// TODO: export this in order to support multiple data sources
type unmarshaller interface {
	// Unmarshal reads configuration data from some source into a struct.
	Unmarshal(dest interface{}) error
}

// Construct an unmarshaller for the given data source.
//
// TODO: export this in order to support multiple data sources
func new(source Getter) unmarshaller {
	return &flatpack{source}
}

// Unexported implementation class for unmarshaller.
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
	case reflect.Ptr:
		// Handle pointers by allocating if necessary, then recursively calling
		// ourselves.
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		err = f.read(name, value.Elem())
	default:
		err = fmt.Errorf("invalid value for %v;  unsupported type %v", name, value.Type())
	}

	return err
}
