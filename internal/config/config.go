package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"

	"log"
	"os"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	AliasLength int64  `yaml:"alias_length" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	PGSQL       `yaml:"pgsql"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8082" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type PGSQL struct {
	DBHost    string `yaml:"db_host" env-default:"localhost"`
	DBPort    int    `yaml:"db_port" env-default:"5432"`
	DBUser    string `yaml:"db_user" env-default:"admin"`
	DBPass    string `yaml:"db_pass" env-default:"secret"`
	DBName    string `yaml:"db_name" env-default:"url_shortener"`
	DBSSLMode string `yaml:"db_ssl_mode" env-default:"disable"`
}

func MustLoad() *Config {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env")
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH нет такой переменной в env")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("файла конфига нет по указанному пути: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("невозможно прочитать конфиг: %s", err)
	}

	return &cfg
}
