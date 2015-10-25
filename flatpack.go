package flatpack

import (
	"fmt"
	"reflect"
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

func Unmarshal(data Getter, v interface{}) error {
	return unmarshal(data, []string{}, v)
}

func unmarshal(data Getter, prefix []string, v interface{}) error {
	vt := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)

	// prepare field-name that we can reuse across fields
	name := make([]string, len(prefix) + 1)
	copy(name, prefix)

	if vt.Kind() != reflect.Struct {
		return fmt.Errorf("Invalid kind for %v: expected struct, got %s", prefix, vt.Kind().String())
	}

	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)
		value := vv.Field(i)

		name[len(name) - 1] = field.Name
		// TODO indirect value if it's not a pointer?
		err := read(data, name, field.Type, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func read(data Getter, name []string, typ reflect.Type, value reflect.Value) error {
	switch typ.Kind() {
	case reflect.Struct:
		unmarshal(data, name, value.Interface())
	default:
		fmt.Println("Reading", name, "which is a", typ.Kind().String())
	}

	return nil
}
