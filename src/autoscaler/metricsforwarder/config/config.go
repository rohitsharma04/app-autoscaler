package config

import (
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServerPort        int               `yaml:"server_port"`
	LogLevel          string            `yaml:"log_level"`
	MetronAddress     string            `yaml:"metron_address"`
	LoggregatorConfig LoggregatorConfig `yaml:"loggregator"`
}

type LoggregatorConfig struct {
	CACertFile     string `yaml:"ca_cert"`
	ClientCertFile string `yaml:"client_cert"`
	ClientKeyFile  string `yaml:"client_key"`
}

const (
	defaultLogLevel      = "info"
	defaultMetronAddress = "127.0.0.1:3458"
	defaultServerPort    = 6110
)

func LoadConfig(reader io.Reader) (*Config, error) {
	conf := &Config{
		ServerPort:    defaultServerPort,
		LogLevel:      defaultLogLevel,
		MetronAddress: defaultMetronAddress,
	}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
