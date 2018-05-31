package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

var config struct {
	HTTP struct {
		ListenPort int    `toml:"listen_port"`
		Domain     string `toml:"domain"`
	} `toml:"http"`
	Database struct {
		Address  string `toml:"address"`
		Name     string `toml:"name"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"database"`
	SigninAPI struct {
		Google struct {
			ClientID     string `toml:"client_id"`
			ClientSecret string `toml:"client_secret"`
		} `toml:"google"`
		Twitter struct {
			APIKey    string `toml:"api_key"`
			APISecret string `toml:"api_secret"`
		} `toml:"twitter"`
		Github struct {
			ClientID     string `toml:"client_id"`
			ClientSecret string `toml:"client_secret"`
		} `toml:"github"`
	} `toml:"signin_api"`
	Misc struct {
		ReservedSlacks []string `toml:"reserved_slacks"`
	} `toml:"misc"`
}

func init() {
	_, err := toml.DecodeFile("./config.toml", &config)
	if err != nil {
		log.Fatalf("Failed to decode toml file '%s': %v", "./config.toml", err)
	}

	initTwitterSignIn(config.SigninAPI.Twitter.APIKey, config.SigninAPI.Twitter.APISecret)
}
