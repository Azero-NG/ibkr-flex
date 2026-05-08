package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvToken      = "IBKR_FLEX_TOKEN"
	EnvQueryID    = "IBKR_FLEX_QUERY_ID"
	EnvConfigPath = "IBKR_FLEX_CONFIG"
)

type Config struct {
	Token   string
	QueryID string
	Source  string // "env", "<config-path>", or "env+<config-path>"
}

var ErrMissing = errors.New("config: required value not set")

// Load resolves the token and query ID, preferring env vars; for any value
// not set in the environment, it falls back to a dotenv-style config file
// (default ~/.config/ibkr-flex/config; override with IBKR_FLEX_CONFIG).
func Load() (*Config, error) {
	cfg := &Config{
		Token:   os.Getenv(EnvToken),
		QueryID: os.Getenv(EnvQueryID),
	}
	envFilled := cfg.Token != "" && cfg.QueryID != ""
	src := []string{}
	if cfg.Token != "" || cfg.QueryID != "" {
		src = append(src, "env")
	}

	if !envFilled {
		path, err := configFilePath()
		if err == nil {
			fileVals, ferr := loadConfigFile(path)
			if ferr == nil {
				if cfg.Token == "" {
					cfg.Token = fileVals[EnvToken]
				}
				if cfg.QueryID == "" {
					cfg.QueryID = fileVals[EnvQueryID]
				}
				if len(fileVals) > 0 {
					src = append(src, path)
				}
			} else if !errors.Is(ferr, fs.ErrNotExist) {
				return nil, fmt.Errorf("config: read %s: %w", path, ferr)
			}
		}
	}

	if cfg.Token == "" {
		return nil, &MissingError{Var: EnvToken}
	}
	if cfg.QueryID == "" {
		return nil, &MissingError{Var: EnvQueryID}
	}
	cfg.Source = strings.Join(src, "+")
	return cfg, nil
}

// DefaultConfigPath returns the default config file path
// (${XDG_CONFIG_HOME:-$HOME/.config}/ibkr-flex/config), or the value of
// IBKR_FLEX_CONFIG when set.
func DefaultConfigPath() (string, error) {
	return configFilePath()
}

func configFilePath() (string, error) {
	if p := os.Getenv(EnvConfigPath); p != "" {
		return p, nil
	}
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ibkr-flex", "config"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ibkr-flex", "config"), nil
}

func loadConfigFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

type MissingError struct {
	Var string
}

func (e *MissingError) Error() string {
	return fmt.Sprintf("config: required value not set: %s (set as env var or write to %s%s)",
		e.Var, defaultPathHint(), missingHelpFor(e.Var))
}

func (e *MissingError) Is(target error) bool {
	return target == ErrMissing
}

func defaultPathHint() string {
	p, err := configFilePath()
	if err != nil {
		return "config file"
	}
	return p
}

func missingHelpFor(varName string) string {
	switch varName {
	case EnvToken:
		return " — see docs/flex-setup.md for IBKR backend steps"
	case EnvQueryID:
		return " — Activity Flex Query ID from IBKR portal"
	}
	return ""
}
