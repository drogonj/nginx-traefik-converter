package configs

type Options struct {
	ProxyBufferHeuristic bool
}

func NewOptions() *Options {
	return &Options{}
}
