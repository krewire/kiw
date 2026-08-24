package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/libs/config"
	"github.com/krewire/libs/core"
)

// ResolveEnv resolves the effective environment with strict precedence:
// override (--env flag) > KIW_ENV > krewire.yaml > local
// (KWL-K4T7W KWL-ENVV-004). An invalid resolved value is a usage error.
func (c *Config) ResolveEnv(override string) (core.Env, error) {
	v := firstNonEmpty(override, os.Getenv("KIW_ENV"))
	if v == "" && c != nil {
		v = c.Env
	}
	return core.ParseEnv(v)
}

// ResolveDebug resolves the effective debug switch with the same precedence:
// an explicitly set override wins, then KIW_DEBUG ("1"/"true"/"yes"), then
// krewire.yaml, then false.
func (c *Config) ResolveDebug(override string, overrideSet bool) bool {
	if overrideSet {
		return isTruthy(override)
	}
	switch strings.ToLower(os.Getenv("KIW_DEBUG")) {
	case "1", "true", "yes":
		return true
	}
	return c != nil && c.Debug
}

// LoadDotEnv parses .env at the project root into the process environment
// without overwriting existing variables (KWN-Q7X4M KWL-CFGV-006).
func (c *Config) LoadDotEnv(root string) error {
	return config.LoadDotEnv(filepath.Join(root, ".env"))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func isTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
