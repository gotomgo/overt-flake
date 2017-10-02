package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// serverConfig is a yaml file representation of the configuration for an
// overt-flake ID server
type serverConfig struct {
	IPAddr       string `yaml:"ipAddr"`
	Epoch        int64  `yaml:"epoch"`
	HidType      string `yaml:"hidType"`
	GenType      string `yaml:"genType"`
	AuthToken    string `yaml:"authToken"`
	HardwareID   []byte `yaml:"hardwareId"`
	MachineID    int64  `yaml:"machineId"`
	DataCenterID int64  `yaml:"dataCenterId"`
}

// loadConfig loads bytes from a file and calls a function to
// translate the bytes into an object
func loadConfig(configPath string, f func([]byte) (interface{}, error)) (interface{}, error) {
	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	config, err := f(configFile)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// loadServerConfig loads an overt-flake server configuration from a yaml file
func loadServerConfig(configPath string) (*serverConfig, error) {
	c, err := loadConfig(configPath, func(bytes []byte) (interface{}, error) {
		var config serverConfig
		err := yaml.Unmarshal(bytes, &config)
		if err != nil {
			return nil, err
		}
		return &config, nil
	})

	if err != nil {
		return nil, err
	}

	return c.(*serverConfig), nil
}

// saveFile saves bytes to a file with owner RW permissions, and R for group and all users
func saveFile(path string, configBytes []byte) error {
	return ioutil.WriteFile(path, configBytes, 0644)
}

// saveServerConfig is a helper function used to save a serverConfig
func saveServerConfig(path string, config *serverConfig) error {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return saveFile(path, bytes)
}
