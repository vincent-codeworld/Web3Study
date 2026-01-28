package config

var Config GlobalConfig

type GlobalConfig struct {
	PostgRestConfig *PostgRestConfig `yaml:"postgres"`
}
type PostgRestConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
	Port     int    `yaml:"port"`
	SslMode  string `yaml:"sslmode"`
	TimeZone string `yaml:"TimeZone"`
}
