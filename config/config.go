package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type Config struct {
	Listen      string    `yaml:"listen"`
	ProbePath   string    `yaml:"probe_path"`
	MetricsPath string    `yaml:"metrics_path"`
	Timeout     float64   `yaml:"timeout"`
	Devices     []*Device `yaml:"devices"`
	Global      Global    `yaml:"global"`

	deviceMap map[string]*Device
}

func DefaultConfig() Config {
	return Config{
		Listen:      ":9777",
		ProbePath:   "/probe",
		MetricsPath: "/metrics",
		Timeout:     60,
		Global: Global{
			Options: DefaultOptions(),
		},
		deviceMap: make(map[string]*Device),
	}
}

func DefaultOptions() Options {
	return Options{
		ExportOLT:      true,
		ExportONUs:     true,
		ExportMACTable: false,
	}
}

func (c *Config) GetDevice(name string) (*Device, bool) {
	d, found := c.deviceMap[name]
	return d, found
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig()

	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	for _, device := range c.Devices {
		if device.Username == nil {
			device.Username = &c.Global.Username
		}
		if device.Password == nil {
			device.Password = &c.Global.Password
		}
		if device.Options == nil {
			device.Options = &c.Global.Options
		}
	}

	if err := c.populateDeviceMap(); err != nil {
		return nil
	}

	return nil
}

func (c *Config) populateDeviceMap() error {
	log.Logger.Trace().Msg("populate target map")
	for _, device := range c.Devices {
		_, exists := c.deviceMap[device.Name]
		if exists {
			return fmt.Errorf("non-unique target name: %s", device.Name)
		}
		log.Logger.Trace().Str("name", device.Name).Msg("add device")
		c.deviceMap[device.Name] = device
	}
	return nil
}

type Global struct {
	Username string  `yaml:"username"`
	Password string  `yaml:"password"`
	Options  Options `yaml:"options"`
}

type Options struct {
	ExportOLT      bool `yaml:"export_olt"`
	ExportONUs     bool `yaml:"export_onus"`
	ExportMACTable bool `yaml:"export_mac_table"`
}

func (o *Options) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*o = DefaultOptions()

	type plain Options
	if err := unmarshal((*plain)(o)); err != nil {
		return err
	}

	return nil
}

type Device struct {
	Name     string   `yaml:"name"`
	Address  string   `yaml:"address"`
	Username *string  `yaml:"username"`
	Password *string  `yaml:"password"`
	Options  *Options `yaml:"options"`
}
