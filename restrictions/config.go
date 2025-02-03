package restrictions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Despire/dnd/atomicfile"
)

type Config struct {
	// LastCommited is the latest commited config
	LastCommited *Config `json:"LastCommited,omitempty"`
	// Version of the config
	Version int64
	// Currently stored restrictions.
	Restrictions map[Type]List
}

func ReadConfig() (*Config, error) {
	if _, err := os.Stat(ConfigPath()); errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	contents, err := os.ReadFile(ConfigPath())
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
	if err := atomicfile.Write(ConfigPath(), b, os.ModePerm); err != nil {
		return fmt.Errorf("failed to atomically write config: %w", err)
	}
	return nil
}

func ConfigPath() string { return fmt.Sprintf("%s/.dnd_config/.config", home) }

func CreateConfigDir() error {
	dir := filepath.Dir(ConfigPath())
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(filepath.Dir(ConfigPath()), os.ModePerm)
	}
	return nil
}
