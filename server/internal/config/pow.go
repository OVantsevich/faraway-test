package config

// Pow - config for proof of work in protocol.
type Pow struct {
	TargetBits  uint8 `env:"TARGET_BITS,notEmpty" envDefault:"0"`
	ReadTimeout int64 `env:"READ_TIMEOUT,notEmpty" envDefault:"60000"`
}
