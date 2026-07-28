package observabilityquery

const (
	DefaultProvider = "greptimedb"
)

type Config struct {
	Provider string `yaml:"provider" json:"provider"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	Database string `yaml:"database" json:"database"`
	Timeout  string `yaml:"timeout" json:"timeout"`
}

func (c Config) Normalize() Config {
	if c.Provider == "" {
		c.Provider = DefaultProvider
	}
	if c.Database == "" {
		c.Database = "public"
	}
	if c.Timeout == "" {
		c.Timeout = "10s"
	}
	return c
}

func (c Config) Configured() bool {
	return c.Endpoint != ""
}
