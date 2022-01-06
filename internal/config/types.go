package config

type PluginConfig struct {
	Name    string   `yaml:"name"`
	Path    string   `yaml:"path"`
	Active  bool     `yaml:"active"`
	Symbols []string `yaml:"symbols"`
}

type Config struct {
	Plugins []PluginConfig `yaml:"plugins"`
}
