package flatpack

type Unmarshaller interface {
	// Unmarshal reads configuration data from some source into a struct.
	Unmarshal(dest interface{}) error
}
