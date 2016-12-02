package flatpack

import (
	"fmt"
	"reflect"
)

// BadType is an error that provides information about invalid types
// encountered while unmarshalling.
type BadType struct {
	Key
	Kind   reflect.Kind
	reason string
}

func (e *BadType) Error() string {
	return fmt.Sprintf("flatpack: invalid type; %s (key=%s,kind=%s)", e.reason, e.Key, e.Kind)
}

// BadValue is an error that provides information about malformed values
// encountered while unmarshalling.
type BadValue struct {
	Key
	Cause    error
	expected string
}

func (e *BadValue) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf(`flatpack: malformed value; (key=%s,cause="%s")`, e.Key, e.Cause.Error())
	}
	return fmt.Sprintf("flatpack: invalid value; expected %s (key=%s)", e.expected, e.Key)
}

// NoReflection is an error that indicates something went wrong when reflecting
// on an unmarshalling target.
type NoReflection struct {
	Key
}

func (e *NoReflection) Error() string {
	return fmt.Sprintf("flatpack: reflection error; unexported field or type? (key=%s)", e.Key)
}
