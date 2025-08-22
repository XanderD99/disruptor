package ffmpeg

// DefaultConfig returns a new Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Exec:       Exec,
		SampleRate: SampleRate,
		Channels:   Channels,
		BufferSize: BufferSize,
	}
}

// Config is used to configure a ffmpeg audio source.
type Config struct {
	Exec       string
	SampleRate int
	Channels   int
	BufferSize int
}
