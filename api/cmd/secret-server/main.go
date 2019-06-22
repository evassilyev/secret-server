package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

func main() {
	var err error

	configPath := pflag.StringP("config", "c", "", "Config file")
	pflag.Parse()

	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if *configPath != "" {
		viper.SetConfigFile(*configPath)
	}

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = ListenAndServe(viper.GetViper())
	if err != nil {
		log.Fatal(err)
	}
}
