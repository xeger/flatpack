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
	_, err := f.unmarshal(Key{}, dest)
	return err
}

// Read configuration source into a struct or sub-struct. Return the number of
// fields that were set.
func (f implementation) unmarshal(prefix Key, dest interface{}) (int, error) {
	count := 0
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, &BadValue{Key: prefix, expected: "non-nil pointer to struct"}
		}
		v = v.Elem()
	} else {
		return 0, &BadType{Key: prefix, Kind: v.Kind(), reason: "expected pointer to struct"}
	}

	vt := v.Type()

	if vt.Kind() != reflect.Struct {
		return 0, &BadType{Key: prefix, Kind: vt.Kind(), reason: "expected struct"}
	}

	// prepare a reusable key whose last element will change as we iterate
	// through the fields in this struct
	name := make(Key, len(prefix)+1)
	copy(name, prefix)

	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)
		value := v.Field(i)

		name[len(name)-1] = field.Name
		read, err := f.read(name, value)
		if err != nil {
			return 0, err
		}
		count += read
	}

	validater, ok := dest.(Validater)
	if ok {
		return count, validater.Validate()
	}

	return count, nil
}

// Coerce a string to a suitable Type and then assign it to a Value (either a
// struct field or an element of a slice).
func (f implementation) assign(dest reflect.Value, source string, name Key) (err error) {
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
				err = &BadValue{Key: name, Cause: numError}
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
				err = &BadValue{Key: name, Cause: numError}
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
				err = &BadValue{Key: name, Cause: numError}
			}
		}
	case reflect.String:
		if err == nil {
			dest.SetString(source)
		}
	default:
		// should be unreachable due to validation in read()
		panic(fmt.Errorf("flatpack: unreachable code in assign(); bug in read()? (kind=%s)", kind))
	}

	return
}

// Set a single struct field by reading a string from the Getter, massaging it
// to the correct Type for that field, and assigning to the given Value.
//
// If the field is itself a struct, recursively unmarshal into the sub-
// struct. If the field is a pointer to anything, allocate it and then
// recursively read into the pointed-to value.
//
// Return the number of fields that were set.
func (f implementation) read(name Key, value reflect.Value) (int, error) {
	count := 0
	vt := value.Type()
	kind := vt.Kind()

	var got string
	var err error

	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64,
		reflect.String:
		got, err = f.source.Get(name)
		if err == nil && got != "" {
			err = f.assign(value, got, name)
			count++
		}
	case reflect.Slice:
		got, err = f.source.Get(name)
		if err == nil && got != "" {
			var raw []interface{}
			err = json.Unmarshal([]byte(got), &raw)
			if err == nil {
				vte := value.Type().Elem()
				value.Set(reflect.MakeSlice(vt, len(raw), len(raw)))
				for i, elem := range raw {
					if err == nil {
						vi := value.Index(i)
						if vte.Kind() == reflect.Ptr {
							vi.Set(reflect.New(vte.Elem()))
							vi = vi.Elem()
						}
						err = f.assign(vi, fmt.Sprintf("%v", elem), name)
						count++
					}
				}
			}
		}
	case reflect.Struct:
		addr := value.Addr()
		if addr.CanInterface() {
			count, err = f.unmarshal(name, addr.Interface())
		} else {
			err = &NoReflection{Key: name}
		}
	case reflect.Ptr:
		// Handle pointers by allocating if necessary, then recursively calling
		// ourselves.
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		count, err = f.read(name, value.Elem())
		// Set pointer (back) to nil if no values were read into it; prevent
		// fooling client into thinking he got nested values when he did not.
		if count == 0 {
			value.Set(reflect.Zero(value.Type()))
		}
	default:
		err = &BadType{Key: name, Kind: value.Kind(), reason: "unsupported data type"}
	}

	return count, err
}
