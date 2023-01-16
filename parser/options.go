package parser

// Option applies an option to the Options
type Option func(*Options)

// Options for the parser
type Options struct {
	// Path to the FoundryVTT world data directory
	Path string
	// Verbose logging
	Verbose bool
}

// NewOptions returns the initialised Options
func NewOptions(opt ...Option) Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}

// Path sets the path to the FoundryVTT world data directory
func Path(a string) Option {
	return func(o *Options) {
		o.Path = a
	}
}

// Verbose sets the verbose option
func Verbose(a bool) Option {
	return func(o *Options) {
		o.Verbose = a
	}
}
