package main

// use https://yaml.to-go.online/ for generation

// Config is the main app config
type Config struct {
	SMTP SMTPConfig `yaml:"smtp"`
}

type SMTPConfig struct {
	Server   string `yaml:"server,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	From     string `yaml:"from,omitempty"`
}
