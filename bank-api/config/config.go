package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/subosito/gotenv"
)

type Config struct {
}

func New() (*Config, error) {
	c := &Config{}
	err := loadEnvFileIfAvailable()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func loadEnvFileIfAvailable() error {
	_, thisFile, _, _ := runtime.Caller(0)
	envFilePath := filepath.Join(filepath.Dir(thisFile), "./env/.env")

	if _, err := os.Stat(envFilePath); err == nil {
		err := gotenv.Load(envFilePath)
		if err != nil {
			return fmt.Errorf("load env files: %w", err)
		}
	}
	return nil
}
