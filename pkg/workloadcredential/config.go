package workloadcredential

import "time"

const (
	DefaultCredentialTTL = 5 * time.Minute
	DefaultClockSkew     = 30 * time.Second
	DefaultBundleTTL     = 10 * time.Minute
)

type KeyState string

const (
	KeyStateActive     KeyState = "ACTIVE"
	KeyStateVerifyOnly KeyState = "VERIFY_ONLY"
)

// Config contains only signing-key references. Private key material must never
// be embedded in YAML or persisted in the control-plane database.
type Config struct {
	Enabled        bool          `yaml:"enabled"`
	Issuer         string        `yaml:"issuer"`
	Audience       string        `yaml:"audience"`
	TrustDomain    string        `yaml:"trustDomain"`
	TTL            time.Duration `yaml:"ttl"`
	ClockSkew      time.Duration `yaml:"clockSkew"`
	BundleSequence uint64        `yaml:"bundleSequence"`
	BundleTTL      time.Duration `yaml:"bundleTTL"`
	Keys           []KeyConfig   `yaml:"keys"`
}

type KeyConfig struct {
	ID             string   `yaml:"id"`
	State          KeyState `yaml:"state"`
	PrivateKeyFile string   `yaml:"privateKeyFile"`
	PublicKeyFile  string   `yaml:"publicKeyFile"`
}

func (c Config) withDefaults() Config {
	if c.TTL == 0 {
		c.TTL = DefaultCredentialTTL
	}
	if c.ClockSkew == 0 {
		c.ClockSkew = DefaultClockSkew
	}
	if c.BundleTTL == 0 {
		c.BundleTTL = DefaultBundleTTL
	}
	return c
}
