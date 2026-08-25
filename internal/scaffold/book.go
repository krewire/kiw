// Package-level file for the manuscript book (VariantBook): kiw init
// --book scaffolds a manuscript/ directory assembled by mdbind (KWM-FX9H2).
package scaffold

import (
	"fmt"
)

// equipBook shapes the kernel into a markdown content book. The sample
// chapters live under content/ (shared with the ssg layout) and the kernel
// main.go is removed, matching the mdbind build path.
func equipBook(opts EquipOptions) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("equip book: project name is required")
	}
	files := []file{
		{krewireYaml, bookKrewireYamlTemplate(opts.Name, opts.Title)},
		{"content/docs/01-introduction.md", introSample},
		{"content/docs/02-getting-started.md", gettingStartedSample},
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
  input: content
  version: 0.1.0
`, name, title)
}

const introSample = "# Introduction\n\nWelcome to your Krewire site.\n\nThis content directory is assembled by **mdbind** into a book-shaped docs website.\n"

const gettingStartedSample = "# Getting Started\n\nWrite Markdown files in the `content/` directory — subdirectories become chapters with subchapters.\n\n```\n$ kiw build\n$ kiw serve\n```\n\nThen open `.krewire/build/docs/getting-started.html`.\n"
