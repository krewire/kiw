// Package gomod reads minimal information out of go.mod files.
package gomod

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Module describes the module path of a go.mod file and its direct requires.
type Module struct {
	// Path is the module path declared in the go.mod file.
	Path string
	// Requires maps the path of each direct dependency to its version.
	Requires map[string]string
}

// Read parses the go.mod file at path.
func Read(path string) (*Module, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m := &Module{Requires: map[string]string{}}
	inRequireBlock := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "module "):
			m.Path = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		case line == "require (":
			inRequireBlock = true
		case line == ")" && inRequireBlock:
			inRequireBlock = false
		case inRequireBlock:
			m.recordRequire(line)
		case strings.HasPrefix(line, "require "):
			m.recordRequire(strings.TrimSpace(strings.TrimPrefix(line, "require ")))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if m.Path == "" {
		return nil, fmt.Errorf("no module directive found in %s", path)
	}
	return m, nil
}

// Find locates the nearest go.mod walking up from dir.
func Find(dir string) (*Module, error) {
	cur, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for {
		p := filepath.Join(cur, "go.mod")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return Read(p)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return nil, fmt.Errorf("go.mod not found (from %s)", dir)
		}
		cur = parent
	}
}

// UsesKrewire reports whether the module depends on the Krewire framework or
// libraries.
func (m *Module) UsesKrewire() bool {
	if m == nil {
		return false
	}
	for path := range m.Requires {
		if strings.HasPrefix(path, "github.com/krewire/") {
			return true
		}
	}
	return false
}

func (m *Module) recordRequire(line string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	m.Requires[fields[0]] = fields[1]
}
