package config

import (
	"flag"
	"os"
)

type Config struct {
	apiPort string
}

func Get() *Config {
	conf := &Config{}

	flag.StringVar(&conf.apiPort, "apiPort", os.Getenv("API_PORT"), "API Port")
	flag.Parse()

	return conf
}

func (c *Config) GetAPIPort() string {
	return ":" + c.apiPort
}
