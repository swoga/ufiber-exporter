package config

type Config struct {
	Listen      string            `yaml:"listen"`
	ProbePath   string            `yaml:"probe_path"`
	MetricsPath string            `yaml:"metrics_path"`
	Timeout     float64           `yaml:"timeout"`
	Devices     map[string]Device `yaml:"devices"`
}

func DefaultConfig() Config {
	return Config{
		Listen:      ":9100",
		ProbePath:   "/probe",
		MetricsPath: "/metrics",
		Timeout:     60,
	}
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig()

	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}

type Device struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
