// Package-level file for the manuscript book (VariantBook): kiw init
// --book scaffolds a manuscript/ directory assembled by mdbind (KWM-FX9H2).
package scaffold

import (
	"fmt"
)

// equipBook shapes the kernel into a manuscript book. The sample chapters
// live under manuscript/ and the kernel main.go is removed, matching the
// mdbind build path.
func equipBook(opts EquipOptions) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("equip book: project name is required")
	}
	files := []file{
		{krewireYaml, bookKrewireYamlTemplate(opts.Name, opts.Title)},
		{"manuscript/01-introduction.md", introSample},
		{"manuscript/02-getting-started.md", gettingStartedSample},
	}
	report, err := writeVariant(opts.Dir, files)
	if err != nil {
		return nil, err
	}
	if err := removeKernelMain(opts.Dir); err != nil {
		return nil, err
	}
	return report, nil
}

func bookKrewireYamlTemplate(name, title string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: book
  title: %s
  author: ""
  base: /
  input: manuscript
  output: site
  version: 0.1.0
`, name, title)
}

const introSample = "# Introduction\n\nWelcome to your Krewire site.\n\nThis manuscript is assembled by **mdbind** into a book-shaped static website.\n"

const gettingStartedSample = "# Getting Started\n\nWrite Markdown files in the `manuscript/` directory.\n\n```\n$ kiw build\n$ kiw serve\n```\n\nThen open `site/index.html`.\n"
