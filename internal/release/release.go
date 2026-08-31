// Package release implements the Krewire ecosystem release topology and the
// version-bump propagation used by `kiw release`.
//
// Every module declares its version and its minimum required versions of other
// modules in `<module>/version.go` (see AGENTS.md, "Version compatibility").
// This package turns a "release module X" decision into a concrete plan of
// source edits: bump X's own Version, and raise every dependent module's
// EcosystemRequires[X] minimum to the new version. It mirrors each module's
// go.mod dependency edges so releases stay mutually compatible.
package release

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krewire/framework"
	"github.com/krewire/guild"
	"github.com/krewire/kiw/internal/version"
	"github.com/krewire/libs/core"
	"github.com/krewire/mdbind"
	"github.com/krewire/ship"
)

// Manifest describes a module's location and version file within the workspace.
// It is the single source of truth for the release topology.
type Manifest struct {
	Name        core.ModuleName
	Dir         string // workspace-relative directory
	VersionFile string // workspace-relative path to the file declaring Version
}

// Modules lists the ecosystem modules in dependency order (dependencies first).
var Modules = []Manifest{
	{Name: core.ModuleLibs, Dir: "libs", VersionFile: "libs/core/version.go"},
	{Name: core.ModuleFramework, Dir: "framework", VersionFile: "framework/version.go"},
	{Name: core.ModuleMdbind, Dir: "mdbind", VersionFile: "mdbind/version.go"},
	{Name: core.ModuleGuild, Dir: "guild", VersionFile: "guild/version.go"},
	{Name: core.ModuleShip, Dir: "ship", VersionFile: "ship/version.go"},
	{Name: core.ModuleKiw, Dir: "kiw", VersionFile: "kiw/internal/version/version.go"},
}

// dependents maps each module to the modules that require it (reverse of go.mod).
var dependents = map[core.ModuleName][]core.ModuleName{
	core.ModuleLibs:      {core.ModuleFramework, core.ModuleMdbind, core.ModuleGuild, core.ModuleShip, core.ModuleKiw},
	core.ModuleFramework: {core.ModuleKiw},
	core.ModuleMdbind:    {core.ModuleKiw},
	core.ModuleGuild:     {core.ModuleKiw},
	core.ModuleShip:      {core.ModuleKiw},
	core.ModuleKiw:       {},
}

// ManifestFor returns the manifest for name.
func ManifestFor(name core.ModuleName) Manifest { return manifest(name) }

func manifest(name core.ModuleName) Manifest {
	for _, m := range Modules {
		if m.Name == name {
			return m
		}
	}
	return Manifest{}
}

func ident(name core.ModuleName) string {
	switch name {
	case core.ModuleLibs:
		return "core.ModuleLibs"
	case core.ModuleFramework:
		return "core.ModuleFramework"
	case core.ModuleMdbind:
		return "core.ModuleMdbind"
	case core.ModuleKiw:
		return "core.ModuleKiw"
	case core.ModuleGuild:
		return "core.ModuleGuild"
	case core.ModuleShip:
		return "core.ModuleShip"
	}
	return ""
}

// CurrentVersion returns the version declared by a module's version.go, read
// from the compiled constant so the plan always reflects the source of truth.
func CurrentVersion(name core.ModuleName) (core.Version, error) {
	switch name {
	case core.ModuleLibs:
		return core.CurrentVersion, nil
	case core.ModuleFramework:
		return framework.Version, nil
	case core.ModuleMdbind:
		return mdbind.Version, nil
	case core.ModuleGuild:
		return guild.Version, nil
	case core.ModuleShip:
		return ship.Version, nil
	case core.ModuleKiw:
		return version.Version, nil
	default:
		return core.Version{}, fmt.Errorf("unknown module %q", name)
	}
}

// RequiredVersion returns the minimum version that dependent currently requires
// of name, read from its compiled EcosystemRequires map.
func RequiredVersion(dependent, name core.ModuleName) (core.Version, bool) {
	switch dependent {
	case core.ModuleLibs:
		return core.Version{}, false
	case core.ModuleFramework:
		v, ok := framework.EcosystemRequires[name]
		return v, ok
	case core.ModuleMdbind:
		v, ok := mdbind.EcosystemRequires[name]
		return v, ok
	case core.ModuleGuild:
		v, ok := guild.EcosystemRequires[name]
		return v, ok
	case core.ModuleShip:
		v, ok := ship.EcosystemRequires[name]
		return v, ok
	case core.ModuleKiw:
		v, ok := version.EcosystemRequires[name]
		return v, ok
	}
	return core.Version{}, false
}

// BumpKind selects which SemVer component to increment.
type BumpKind string

const (
	BumpPatch BumpKind = "patch"
	BumpMinor BumpKind = "minor"
	BumpMajor BumpKind = "major"
)

// ParseBump validates and converts s to a BumpKind.
func ParseBump(s string) (BumpKind, error) {
	switch BumpKind(s) {
	case BumpPatch, BumpMinor, BumpMajor:
		return BumpKind(s), nil
	}
	return "", fmt.Errorf("invalid bump %q: want patch|minor|major", s)
}

// Bump returns v incremented per k.
func Bump(v core.Version, k BumpKind) core.Version {
	switch k {
	case BumpMajor:
		return core.Version{Major: v.Major + 1}
	case BumpMinor:
		return core.Version{Major: v.Major, Minor: v.Minor + 1}
	default: // patch
		return core.Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
	}
}

// Edit is a single planned source change.
type Edit struct {
	Module  core.ModuleName
	File    string // workspace-relative path
	Summary string
	From    string // exact substring to replace (asserted unique in the file)
	To      string
}

// Plan computes the edits for releasing the given modules with the given bump.
// For each released module it bumps its own Version and, for every dependent,
// raises that dependent's EcosystemRequires[released] minimum to the new version.
func Plan(released []core.ModuleName, bump BumpKind) ([]Edit, error) {
	seen := map[core.ModuleName]bool{}
	var edits []Edit
	for _, r := range released {
		if seen[r] {
			continue
		}
		seen[r] = true
		cur, err := CurrentVersion(r)
		if err != nil {
			return nil, err
		}
		nv := Bump(cur, bump)
		m := manifest(r)
		edits = append(edits, Edit{
			Module:  r,
			File:    m.VersionFile,
			Summary: fmt.Sprintf("bump %s %s -> %s", r, cur.String(), nv.String()),
			From:    versionDecl(r, cur.String()),
			To:      versionDecl(r, nv.String()),
		})
		for _, d := range dependents[r] {
			req, ok := RequiredVersion(d, r)
			if !ok {
				continue
			}
			dm := manifest(d)
			edits = append(edits, Edit{
				Module:  d,
				File:    dm.VersionFile,
				Summary: fmt.Sprintf("raise %s requires %s -> %s", d, r, nv.String()),
				From:    reqDecl(r, req.String()),
				To:      reqDecl(r, nv.String()),
			})
		}
	}
	sort.Slice(edits, func(i, j int) bool {
		if edits[i].Module != edits[j].Module {
			return edits[i].Module < edits[j].Module
		}
		return edits[i].File < edits[j].File
	})
	return edits, nil
}

// Apply writes the planned edits to disk under root. When dryRun is true it
// only validates that each edit is applicable (unique match) without writing.
// It returns the list of files that were actually modified.
func Apply(edits []Edit, root string, dryRun bool) ([]string, error) {
	var modified []string
	for _, e := range edits {
		p := filepath.Join(root, e.File)
		data, err := os.ReadFile(p)
		if err != nil {
			return modified, err
		}
		count := strings.Count(string(data), e.From)
		if count == 0 {
			return modified, fmt.Errorf("release: %q not found in %s", e.From, e.File)
		}
		if count > 1 {
			return modified, fmt.Errorf("release: %q is ambiguous (%d matches) in %s", e.From, count, e.File)
		}
		if dryRun {
			continue
		}
		updated := strings.Replace(string(data), e.From, e.To, 1)
		if err := os.WriteFile(p, []byte(updated), 0o644); err != nil {
			return modified, err
		}
		modified = append(modified, e.File)
	}
	return modified, nil
}

func versionDecl(name core.ModuleName, v string) string {
	if name == core.ModuleLibs {
		return fmt.Sprintf("CurrentVersion = MustParseVersion(\"%s\")", v)
	}
	return fmt.Sprintf("var Version = core.MustParseVersion(\"%s\")", v)
}

func reqDecl(name core.ModuleName, v string) string {
	return fmt.Sprintf("%s: core.MustParseVersion(\"%s\")", ident(name), v)
}
