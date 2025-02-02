package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Despire/dnd/atomicfile"
	"github.com/Despire/dnd/restrictions"
)

type Config struct {
	// LastCommited is the latest commited config
	LastCommited *Config `json:"LastCommited,omitempty"`
	// Version of the config
	Version int64
	// Currently stored restrictions.
	Restrictions map[restrictions.Type]restrictions.List
}

func ReadConfig() (*Config, error) {
	if _, err := os.Stat(MustConfigPath()); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	contents, err := os.ReadFile(MustConfigPath())
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(contents, &cfg)
	return &cfg, err
}

func WriteConfig(c *Config) error {
	c = &Config{
		LastCommited: c.LastCommited,
		Version:      c.Version + 1,
		Restrictions: c.Restrictions,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := atomicfile.Write(MustConfigPath(), b, os.ModePerm); err != nil {
		return fmt.Errorf("failed to atomically write config: %w", err)
	}
	return nil
}

func MustConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get retrieve home directory: %v", err))
	}
	return fmt.Sprintf("%s/.dnd_config/.config", home)
}

func CreateConfigDir() error {
	dir := filepath.Dir(MustConfigPath())
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(filepath.Dir(MustConfigPath()), os.ModePerm)
	}
	return nil
}
