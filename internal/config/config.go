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
	Port     int    `env:"POSTGRES_PORT,notEmpty"`
	User     string `env:"POSTGRES_USER,notEmpty"`
	Password string `env:"POSTGRES_PASSWORD,notEmpty"`
	DbName   string `env:"POSTGRES_DB,notEmpty"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string `env:"SERVER_HOST,notEmpty"`
	Port int    `env:"SERVER_PORT,notEmpty"`
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

// LoadConfig loads configuration from .env file or environment variables.
// Uses sync.Once to ensure the config is loaded only once (singleton pattern).
// If .env file exists, loads variables from it. Otherwise, uses OS environment variables.
// Panics if required environment variables are missing or contain invalid values.
func LoadConfig() *Config {
	once.Do(func() {
		absFP, _ := filepath.Abs(filePath)

		// Try to load .env file, but don't fail if it doesn't exist
		if _, err := os.Stat(filePath); err == nil {
			// File exists, try to load it
			err = godotenv.Load(filePath)
			if err != nil {
				log.Fatal(fmt.Sprintf("error loading env file `%s`", absFP), zap.Error(err))
			}
			log.Println(fmt.Sprintf("Loaded environment variables from `%s`", absFP))
		} else if os.IsNotExist(err) {
			// File doesn't exist, use environment variables from Docker/OS
			log.Println("No .env file found, using environment variables from system")
		} else {
			// Some other error occurred
			log.Fatal(fmt.Sprintf("error checking `%s` file", absFP), zap.Error(err))
		}

		// Parse environment variables into Config struct
		cfg = &Config{}
		err := env.ParseNested(cfg)
		if err != nil {
			log.Fatal("error parsing environment variables", zap.Error(err))
		}
	})

	return cfg
}
