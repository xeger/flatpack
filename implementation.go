package flatpack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Unexported implementation class for unmarshaller.
type implementation struct {
	source Getter
}

// Unmarshal reads configuration data from some source into a struct.
func (f implementation) Unmarshal(dest interface{}) error {
	return f.unmarshal(Key{}, dest)
}

// Read configuration source into a struct or sub-struct.
func (f implementation) unmarshal(prefix Key, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("invalid value: need non-nil pointer")
		}
		v = v.Elem()
	} else {
		return fmt.Errorf("invalid type: expected pointer-to-struct (key=%s,type=%s)", prefix, v.Kind())
	}

	vt := v.Type()

	if vt.Kind() != reflect.Struct {
		return fmt.Errorf("invalid type: expected struct (key=%s,type=%s)", prefix, vt.Kind())
	}

	// prepare a reusable key whose last element will change as we iterate
	// through the fields in this struct
	name := make(Key, len(prefix)+1)
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

	validater, ok := dest.(Validater)
	if ok {
		return validater.Validate()
	}

	return nil
}

// Coerce a value to a suitable Type and then assign it to a Value (either a
// struct field or an element of a slice).
func (f implementation) assign(dest reflect.Value, source string) (err error) {
	kind := dest.Type().Kind()

	switch kind {
	case reflect.Bool:
		var boolean bool
		boolean, err = strconv.ParseBool(source)
		if err == nil {
			dest.SetBool(boolean)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var number int64
		number, err = strconv.ParseInt(source, 10, int(dest.Type().Size()*8))
		if err == nil {
			dest.SetInt(number)
		} else {
			numError, ok := err.(*strconv.NumError)
			if ok {
				err = fmt.Errorf("invalid value: cannot parse string as integer (value=%s,error=%s)", numError.Num, numError.Err)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		var number uint64
		number, err = strconv.ParseUint(source, 10, int(dest.Type().Size()*8))
		if err == nil {
			dest.SetUint(number)
		} else {
			numError, ok := err.(*strconv.NumError)
			if ok {
				err = fmt.Errorf("invalid value: cannot parse string as integer: (value=%s,error=%s)", numError.Num, numError.Err)
			}
		}
	case reflect.Float32, reflect.Float64:
		var number float64
		number, err = strconv.ParseFloat(source, int(dest.Type().Size()*8))
		if err == nil {
			dest.SetFloat(number)
		} else {
			numError, ok := err.(*strconv.NumError)
			if ok {
				err = fmt.Errorf("invalid value: cannot parse string as float (value=%s,error=%s)", numError.Num, numError.Err)
			}
		}
	case reflect.String:
		if err == nil {
			dest.SetString(source)
		}
	default:
		// should be unreachable due to validation in read()
		panic("case should be unreachable; bug in flatpack.unmarshaller.read9)")
	}

	return
}

// Set a single struct field by reading a string from the Getter, massaging it
// to the correct Type for that field, and assigning to the given Value.
func (f implementation) read(name Key, value reflect.Value) error {
	kind := value.Type().Kind()

	var got string
	var err error

	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.String:
		got, err = f.source.Get(name)
		if err == nil && got != "" {
			err = f.assign(value, got)
		}
	case reflect.Array, reflect.Slice:
		got, err = f.source.Get(name)
		if err == nil && got != "" {
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
		err = fmt.Errorf("invalid type: unsupported data type (key=%s,type=%s)", name, value.Type())
	}

	return err
}
