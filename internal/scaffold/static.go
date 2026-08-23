// Package-level file for the declarative static site (VariantStatic):
// kiw init --static writes a krewire.yaml whose ssg: key drives web/ssg.
// No ssg.yaml is produced (KWF-PT8OD).
package scaffold

import (
	"fmt"
)

// equipStatic shapes the kernel into a declarative static site. The ssg
// definition lives exclusively in krewire.yaml, and the kernel's placeholder
// main.go is removed so shape detection is pinned by project.kind.
func equipStatic(opts EquipOptions) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("equip static: project name is required")
	}
	files := []file{
		{krewireYaml, staticKrewireYamlTemplate(opts.Name, opts.Title)},
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

func staticKrewireYamlTemplate(name, title string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: site
  title: %s
  description: A static site scaffolded by kiw init.
  output: site
  version: 0.1.0
ssg:
  layouts:
    - name: Base
      body: |-
        <!DOCTYPE html>
        <html lang="en">
        <head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}}</title></head>
        <body><main>{{.Content}}</main></body>
        </html>
  components:
    - name: Hero
      body: |-
        <section><h1>{{.Title}}</h1><p>{{.Description}}</p></section>
  pages:
    - path: /
      title: Home
      layout: Base
      root: Hero
`, name, title)
}
