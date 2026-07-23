package otel

import "time"

type Config struct {
	Endpoint           string        `yaml:"endpoint"`
	LogsEndpoint       string        `yaml:"logsEndpoint"`
	Timeout            time.Duration `yaml:"timeout"`
	ReconnectionPeriod time.Duration `yaml:"reconnectionPeriod"`
	Compressor         string        `yaml:"compressor"`
	Prefix             string        `yaml:"prefix"`
	PushInterval       time.Duration `yaml:"pushInterval"`
	LogQueueSize       int           `yaml:"logQueueSize"`
	LogSpoolDir        string        `yaml:"logSpoolDir"`
	LogSpoolBucket     string        `yaml:"logSpoolBucket"`
	LogSpoolMaxBytes   int64         `yaml:"logSpoolMaxBytes"`
	ServiceName        string        `yaml:"serviceName"`
	ServiceVersion     string        `yaml:"serviceVersion"`
	Environment        string        `yaml:"environment"`
	Cluster            string        `yaml:"cluster"`
	NodeRole           string        `yaml:"nodeRole"`
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
	if c.LogQueueSize <= 0 {
		c.LogQueueSize = 2048
	}
	if c.LogSpoolMaxBytes <= 0 {
		c.LogSpoolMaxBytes = 64 * 1024 * 1024
	}
	if c.ServiceName == "" {
		c.ServiceName = "pole-control-plane"
	}
	if c.Environment == "" {
		c.Environment = "local"
	}
	if c.Cluster == "" {
		c.Cluster = "default"
	}
	if c.NodeRole == "" {
		c.NodeRole = "control-plane"
	}
}
