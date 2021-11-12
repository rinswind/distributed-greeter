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
		Endpoint string `yaml:"Endpoint" env:"DB_ENDPOINT,overwrite"`
		User     string `yaml:"User" env:"DB_USER,overwrite"`
		Password string `yaml:"Password" env:"DB_PASSWORD,overwrite"`
	} `yaml:"Db"`
	DbConfigDir string `yaml:"DbConfigDir" env:"DB_CONFIG_DIR,overwrite"`

	Redis struct {
		Endpoint  string `yaml:"Endpoint" env:"REDIS_ENDPOINT,overwrite"`
		AccessKey string `yaml:"AccessKey" env:"REDIS_ACCESS_KEY,overwrite"`
	} `yaml:"Redis"`
	RedisConfigDir string `yaml:"RedisConfigDir" env:"REDIS_CONFIG_DIR,overwrite"`

	AccessToken struct {
		AccessTokenSecret  string `yaml:"AccessTokenSecret" env:"AT_ACCESS_TOKEN_SECRET",overwrite`
		RefreshTokenSecret string `yaml:"RefreshTokenSecret" env:"AT_REFRESH_TOKEN_SECRET",overwrite`
	} `yaml:"AccessToken"`
	AccessTokenConfigDir string `yaml:"AccessTokenConfigDir" env:"AT_CONFIG_DIR,overwrite"`
}

func ReadConfig() *Config {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "config.yaml"
	}

	var cfg Config

	loadYaml(&cfg, configFile)
	loadEnv(&cfg)

	loadEnv(&cfg.Http)

	loadDir(&cfg.Db, cfg.DbConfigDir)
	loadEnv(&cfg.Db)

	loadDir(&cfg.Redis, cfg.RedisConfigDir)
	loadEnv(&cfg.Redis)

	loadDir(&cfg.AccessToken, cfg.AccessTokenConfigDir)
	loadEnv(&cfg.AccessToken)

	return &cfg
}
