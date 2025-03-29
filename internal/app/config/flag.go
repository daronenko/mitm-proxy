package config

import "flag"

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to app config file")
}
