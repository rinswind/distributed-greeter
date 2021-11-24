package config

import (
	"os"
)

type Config struct {
	Http struct {
		Port     int    `yaml:"Port" env:"HTTP_PORT,overwrite"`
		CertFile string `yaml:"CertFile" env:"HTTP_CERT_FILE,overwrite"`
	} `yaml:"Http"`

	Db struct {
		Dsn      string `yaml:"Dsn" env:"DSN,overwrite"`
		Driver   string `yaml:"Driver" env:"DRIVER,overwite`
		Endpoint string `yaml:"Endpoint" env:"ENDPOINT,overwrite"`
		Name     string `yaml:"Name" env:"NAME,overwrite"`
		User     string `yaml:"User" env:"USER,overwrite"`
		Password string `yaml:"Password" env:"PASSWORD,overwrite"`
	} `yaml:"Db" env:",prefix=DB_"`
	DbConfigDir string `yaml:"DbConfigDir"`

	Redis struct {
		Dsn      string `yaml:"Dsn" env:"DSN,overwrite"`
		Endpoint string `yaml:"Endpoint" env:"ENDPOINT,overwrite"`
		Db       int    `yaml:"Db" env:"DB,overwrite"`
		User     string `yaml:"User" env:"User,overwrite"`
		Password string `yaml:"Password" env:"PASSWORD,overwrite"`
		TLS      bool   `yaml:"TLS" env:"TLS,overwrite"`
	} `yaml:"Redis" env:",prefix=REDIS_"`
	RedisConfigDir string `yaml:"RedisConfigDir"`

	AccessToken struct {
		AccessTokenSecret  string `yaml:"AccessTokenSecret" env:"ACCESS_TOKEN_SECRET",overwrite`
		RefreshTokenSecret string `yaml:"RefreshTokenSecret" env:"REFRESH_TOKEN_SECRET",overwrite`
	} `yaml:"AccessToken" env:",prefix=AT_"`
	AccessTokenConfigDir string `yaml:"AccessTokenConfigDir"`
}

func ReadConfig() *Config {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}

	var cfg Config

	loadYaml(&cfg, configFile)

	loadDir(&cfg, cfg.DbConfigDir)
	loadDir(&cfg, cfg.RedisConfigDir)
	loadDir(&cfg, cfg.AccessTokenConfigDir)

	loadEnv(&cfg)

	expandTemplate(&cfg.Db.Dsn, &cfg.Db)
	expandTemplate(&cfg.Redis.Dsn, &cfg.Redis)

	return &cfg
}
