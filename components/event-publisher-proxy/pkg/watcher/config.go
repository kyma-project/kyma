package watcher

import "flag"

// Config represents environment config.
type Config struct {
	ServerAddress   string
	PublishEndpoint string `config:"publish_endpoint"`
}

func New() *Config {
	c := new(Config)
	flag.StringVar(&c.ServerAddress, "addr", ":8888", "HTTP Server listen address.")
	flag.Parse()
	return c
}
