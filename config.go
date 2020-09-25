package main

var (
	Version           = "0.0.0"
	DefaultConfigName = "config"
	DefaultConfigPath = "."
	EnvPrefix         = ""
)

var Config = struct {
	TitleString string `mapstructure:"title" short:"t" default:"12" desc:"AppTitle"`
	Text        string `mapstructure:"text,omitempty"`
}{}
