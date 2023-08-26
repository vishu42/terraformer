package pkg

import "os"

type Config struct {
	ClientSecret string
	Debug        bool
}

func LoadConfig() Config {
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		panic("CLIENT_SECRET not set")
	}

	debug := os.Getenv("DEBUG")
	return Config{
		ClientSecret: clientSecret,
		Debug:        debug == "true",
	}
}
