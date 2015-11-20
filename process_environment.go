package flatpack

// A getter that reads configuration data from the process environment (or
// something similar).
type processEnvironment struct {
	lookup func(string) (string, bool)
}

func (pe processEnvironment) Get(name Key) (string, error) {
	key := name.AsEnv()
	value, _ := pe.lookup(key)
	return value, nil
}
