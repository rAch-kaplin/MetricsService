package config

const (
	DefaultEndpoint        = "localhost:8080"
	DefaultStoreInterval   = 300
	DefaultFileStoragePath = "/temp/metrics-db.json"
	DefaultRestoreOnStart  = true
)

type Options struct {
	EndPointAddr    string
	StoreInterval   int
	FileStoragePath string
	RestoreOnStart  bool
}

type EnvConfig struct {
	EndPointAddr    string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	RestoreOnStart  bool   `env:"RESTORE"`
}

func NewOptions(envOpts, flagOpts []func(*Options)) *Options {
	opts := &Options{
		EndPointAddr:    DefaultEndpoint,
		StoreInterval:   DefaultStoreInterval,
		FileStoragePath: DefaultFileStoragePath,
		RestoreOnStart:  DefaultRestoreOnStart,
	}

	for _, opt := range flagOpts {
		opt(opts)
	}

	for _, opt := range envOpts {
		opt(opts)
	}

	return opts
}

func WithAddress(addr string) func(*Options) {
	return func(o *Options) {
		o.EndPointAddr = addr
	}
}

func WithStoreInterval(interval int) func(*Options) {
	return func(o *Options) {
		o.StoreInterval = interval
	}
}

func WithFileStoragePath(path string) func(*Options) {
	return func(o *Options) {
		o.FileStoragePath = path
	}
}

func WithRestoreOnStart(restore bool) func(*Options) {
	return func(o *Options) {
		o.RestoreOnStart = restore
	}
}
