package config

import (
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var ROOT_PATH string = func() string {

	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")

		if len(parts) > 1 && parts[0] == "HOME" {
			return parts[1] + "/.gocrawler"
		}
	}

	return ""
}()

var CONFIG_FILE = ROOT_PATH + "/config.yaml"

func GetConfig() *Config {

	source, err := ioutil.ReadFile(CONFIG_FILE)
	if err != nil {
		return nil
	}

	var config *Config
	err = yaml.Unmarshal(source, config)
	if err != nil {
		return nil
	}

	return config
}

func SaveConfig(config Config) bool {

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return false
	}

	err = ioutil.WriteFile(CONFIG_FILE, bytes, 0777)
	return err == nil
}
