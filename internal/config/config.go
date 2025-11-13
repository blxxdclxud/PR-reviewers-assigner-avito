package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gabrielsoaressantos/env/v8"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// DBConfig contains PostgreSQL database connection settings
type DBConfig struct {
	Host     string `env:"POSTGRES_HOST,notEmpty"`
	Port     string `env:"POSTGRES_PORT,notEmpty"`
	User     string `env:"POSTGRES_USER,notEmpty"`
	Password string `env:"POSTGRES_PASSWORD,notEmpty"`
	DbName   string `env:"POSTGRES_DB,notEmpty"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string `env:"SERVER_HOST,notEmpty"`
	Port string `env:"SERVER_PORT,notEmpty"`
}

// Config contains all application config
type Config struct {
	DBConfig
	ServerConfig
}

const filePath = "./.env"

var (
	cfg  *Config
	once sync.Once
)

// LoadConfig loads configuration from .env file.
// Uses sync.Once to ensure the config is loaded only once (singleton pattern).
// Panics if the .env file is missing or contains invalid values.
func LoadConfig() *Config {
	once.Do(func() {
		absFP, _ := filepath.Abs(filePath)

		// Check if .env file exists
		_, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(fmt.Sprintf("error finding `%s` file", absFP), zap.Error(err))
		}

		// Load environment variables from .env
		err = godotenv.Load(filePath)
		if err != nil {
			log.Fatal(fmt.Sprintf("error loading env file `%s`", absFP), zap.Error(err))
		}

		// Parse environment variables into Config struct
		cfg = &Config{}
		err = env.ParseNested(cfg)
		if err != nil {
			log.Fatal(fmt.Sprintf("error parsing config file `%s`", absFP), zap.Error(err))
		}
	})

	return cfg
}
