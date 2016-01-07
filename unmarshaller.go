package flatpack

// unmarshaller represents an object that is capable of unmarshalling
// configuration data into destination structures. It encapsulates the
// source of the data as well as any options pertaining to data ccess.
//
// TODO: export this in order to provide a non-singleton interface to flatpack
type unmarshaller interface {
	// Unmarshal reads configuration data from some source into a struct.
	Unmarshal(dest interface{}) error
}

// Construct an unmarshaller for the given data source.
//
// TODO: export this in order to provide a non-singleton interface to flatpack
func new(source Getter) unmarshaller {
	return &implementation{source}
}
