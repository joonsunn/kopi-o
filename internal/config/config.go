// Package config resolves, loads, and saves kopi-o user config.
//
// Location: $KOPI_O_CONFIG override, else $XDG_CONFIG_HOME/kopi-o/config.json,
// else ~/.config/kopi-o/config.json. Directory is 0700, file is 0600,
// writes are atomic (tmp + rename). Missing or corrupt files yield defaults.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	appDirName  = "kopi-o"
	configName  = "config.json"
	dirPerm     = 0o700
	filePerm    = 0o600
	defaultSecs = 3600
)

// Config is the persisted user config.
type Config struct {
	PresetSecs    int  `json:"presetSecs"`
	Indefinite    bool `json:"indefinite"`
	LaunchAtLogin bool `json:"launchAtLogin"`
}

// Default returns the default config (1h preset).
func Default() Config {
	return Config{PresetSecs: defaultSecs}
}

// Path resolves the config file path.
func Path() string {
	if p := os.Getenv("KOPI_O_CONFIG"); p != "" {
		return p
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			home = os.ExpandEnv("$HOME")
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, appDirName, configName)
}

// Load reads config from Path. Missing file or corrupt JSON yields defaults.
func Load() Config {
	return LoadFrom(Path())
}

// LoadFrom reads config from an explicit path (test hook).
// Missing file or corrupt JSON yields defaults.
func LoadFrom(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default()
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Default()
	}
	if c.PresetSecs <= 0 && !c.Indefinite {
		// Zero-value / empty file: fall back to defaults.
		// A valid indefinite config has Indefinite=true, keep it.
		return Default()
	}
	return c
}

// Save writes c to Path atomically (tmp + rename), 0700 dir, 0600 file.
func Save(c Config) error {
	return SaveTo(Path(), c)
}

// SaveTo writes c to an explicit path atomically (test hook).
func SaveTo(path string, c Config) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Ensure cleanup on failure; success renames away.
	defer os.Remove(tmpName)
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(filePerm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, filePerm); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
