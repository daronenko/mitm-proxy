package logger

type Config struct {
	TraceLevel bool   `mapstructure:"traceLevel"`
	StdoutOnly bool   `mapstructure:"stdoutOnly"`
	Path       string `mapstructure:"path"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxAge     int    `mapstructure:"maxAge"`
	MaxBackups int    `mapstructure:"maxBackups"`
	Compress   bool   `mapstructure:"compress"`
}
