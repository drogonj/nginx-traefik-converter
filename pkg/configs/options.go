package configs

// Options holds the options required to run the converters.
type Options struct {
	ProxyBufferHeuristic bool
}

// NewOptions returns new instance of Options when invoked.
func NewOptions() *Options {
	return &Options{}
}
