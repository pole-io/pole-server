package otel

import "time"

type Config struct {
	Endpoint           string        `yaml:"endpoint"`
	Timeout            time.Duration `yaml:"timeout"`
	ReconnectionPeriod time.Duration `yaml:"reconnectionPeriod"`
	Compressor         string        `yaml:"compressor"`
	Prefix             string        `yaml:"prefix"`
	PushInterval       time.Duration `yaml:"pushInterval"`
}

func (c *Config) setDefault() {
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	if c.ReconnectionPeriod == 0 {
		c.ReconnectionPeriod = 10 * time.Second
	}
	if c.Prefix == "" {
		c.Prefix = "wns_"
	}
	if c.Compressor == "" {
		c.Compressor = "gzip"
	}
	// 默认推送间隔为 5 秒
	if c.PushInterval == 0 {
		c.PushInterval = 5 * time.Second
	}
}
