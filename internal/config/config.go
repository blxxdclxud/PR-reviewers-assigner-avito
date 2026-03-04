package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/gabrielsoaressantos/env/v8"
	"github.com/joho/godotenv"
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
		absFP, err := filepath.Abs(filePath)
		if err != nil {
			log.Fatalf("failed to resolve path %q: %v", filePath, err)
		}

		// Try to load .env file, but don't fail if it doesn't exist
		if _, err := os.Stat(filePath); err == nil {
			// File exists, try to load it
			err = godotenv.Load(filePath)
			if err != nil {
				log.Fatalf("error loading env file %q: %v", absFP, err)
			}
			log.Printf("Loaded environment variables from %q", absFP)
		} else if os.IsNotExist(err) {
			// File doesn't exist, use environment variables from Docker/OS
			log.Println("No .env file found, using environment variables from system")
		} else {
			// Some other error occurred
			log.Fatalf("error checking file %q: %v", absFP, err)
		}

		// Parse environment variables into Config struct
		cfg = &Config{}
		if err = env.ParseNested(cfg); err != nil {
			log.Fatalf("error parsing environment variables: %v", err)
		}
	})

	return cfg
}
