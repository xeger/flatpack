package flatpack

import "os"

// Getter represents a read-only repository of key/value pairs where the keys
// are ordered sequences of strings and the values are strings. It's analogous
// to a map[Key]string, but the data may be retrieved from a network or
// filesystem source and the structured key name may be treated as an indicator
// of hierarchy or containment within the data source, e.g. URL hierarchy on an
// HTTP k/v store, or directory hierarchy on a filesystem-based store.
type Getter interface {
	Get(name Key) (string, error)
}

// Validater represents an object that knows how to validate itself. If the
// object you pass to Unmarshal implements this interface, flatpack will call
// it for you and return the error if anything fails to validate.
//
// Note that this is a ValidatER (a thing that can Validate itself), not
// a ValidatOR (a thing that knows how to validate other things).
type Validater interface {
	Validate() error
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
// NOTE: flatpack currently lacks a public non-singleton interface, but it
// would be easy to add by simply exporting flatpack.new() and
// declaring an Unmarshaller interface.
func Unmarshal(dest interface{}) error {
	return new(DataSource).Unmarshal(dest)
}
