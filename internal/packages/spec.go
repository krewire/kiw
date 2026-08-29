package packages

import (
	"fmt"
	"strings"
)

// Spec is a parsed package@version reference.
type Spec struct {
	Raw     string // original input e.g. "twcss@1.2.3" or "pkg@latest"
	Name    string // package name e.g. "twcss" or "@scope/pkg" or "github.com/foo/bar"
	Version string // version e.g. "1.2.3", "latest", "" (empty means latest)
}

// ParseSpec parses raw package@version. It supports:
//   - "pkg"                → {Name:"pkg", Version:""}
//   - "pkg@1.2.3"           → {Name:"pkg", Version:"1.2.3"}
//   - "pkg@latest"          → {Name:"pkg", Version:"latest"}
//   - "@scope/pkg@1.0.0"    → {Name:"@scope/pkg", Version:"1.0.0"}
//   - "@scope/pkg"          → {Name:"@scope/pkg", Version:""}
//   - "github.com/foo/bar@v1.2.3" → {Name:"github.com/foo/bar", Version:"v1.2.3"}
func ParseSpec(raw string) (Spec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Spec{}, fmt.Errorf("empty package spec")
	}
	// Scoped npm package: starts with @
	if strings.HasPrefix(raw, "@") {
		// find second @ after scope/name
		slash := strings.Index(raw, "/")
		if slash == -1 {
			return Spec{}, fmt.Errorf("invalid scoped package %q", raw)
		}
		rest := raw[slash+1:]
		if at := strings.LastIndex(rest, "@"); at != -1 {
			name := raw[:slash+1+at]
			ver := rest[at+1:]
			if ver == "" {
				return Spec{}, fmt.Errorf("invalid version in %q", raw)
			}
			return Spec{Raw: raw, Name: name, Version: ver}, nil
		}
		return Spec{Raw: raw, Name: raw, Version: ""}, nil
	}
	// Non-scoped: split at last @ (to handle github.com/foo/bar@v1.2.3)
	if at := strings.LastIndex(raw, "@"); at != -1 {
		name := raw[:at]
		ver := raw[at+1:]
		if name == "" || ver == "" {
			return Spec{}, fmt.Errorf("invalid package spec %q", raw)
		}
		return Spec{Raw: raw, Name: name, Version: ver}, nil
	}
	return Spec{Raw: raw, Name: raw, Version: ""}, nil
}

// EffectiveVersion returns version to use for install ("" or "latest" → "latest").
func (s Spec) EffectiveVersion() string {
	if s.Version == "" {
		return "latest"
	}
	return s.Version
}

// String returns canonical package@version (omits version if latest/empty).
func (s Spec) String() string {
	if s.Version == "" || s.Version == "latest" {
		return s.Name + "@latest"
	}
	return s.Name + "@" + s.Version
}
