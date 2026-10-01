package commands

import (
	"strings"
	"testing"

	"github.com/krewire/framework/web/ssg"
	"github.com/krewire/mdbind/book"
)

func TestVersionedAssetURLs(t *testing.T) {
	cases := []struct {
		name    string
		urls    []string
		version string
		want    []string
	}{
		{"stamps version", []string{"/assets/a.css"}, "0.1.0", []string{"/assets/a.css?v=0.1.0"}},
		{"strips v prefix", []string{"/assets/a.css"}, "v1.2.3", []string{"/assets/a.css?v=1.2.3"}},
		{"no version leaves urls untouched", []string{"/assets/a.css"}, "", []string{"/assets/a.css"}},
		{"empty urls", nil, "1.0.0", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := versionedAssetURLs(c.urls, c.version)
			if len(got) != len(c.want) {
				t.Fatalf("got %d urls %v, want %d %v", len(got), got, len(c.want), c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("url[%d] = %q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestWithoutAssetDropsMatchingEntry(t *testing.T) {
	plan := []string{"/assets/style.css", "/assets/tailwind.css", "/assets/mdbind.css"}
	got := withoutAsset(plan, book.StylesheetName)
	if len(got) != 2 {
		t.Fatalf("got %v, want 2 entries", got)
	}
	for _, u := range got {
		if strings.Contains(u, "mdbind.css") {
			t.Errorf("mdbind.css must be dropped, got %q", u)
		}
	}
}

// TestBookStylesheetIsPartOfTheSitePlan pins the coupling between mdbind and
// framework/web/ssg: the site's asset plan lists the book's stylesheet, and the
// devtool filters it out because mdbind links it itself. If either side renames
// it, this test fails instead of the page loading the stylesheet twice.
func TestBookStylesheetIsPartOfTheSitePlan(t *testing.T) {
	s := ssg.New().
		Asset("assets/style.css", "").
		Asset(book.StylesheetName, "")
	css, _ := s.InjectedAssets()
	found := false
	for _, u := range css {
		if strings.HasSuffix(u, book.StylesheetName) {
			found = true
		}
	}
	if !found {
		t.Fatalf("site plan %v does not contain %q", css, book.StylesheetName)
	}
	if got := withoutAsset(css, book.StylesheetName); len(got) != len(css)-1 {
		t.Errorf("filtering removed %d entries, want 1", len(css)-len(got))
	}
}

func TestWithoutAssetKeepsUnrelatedEntries(t *testing.T) {
	plan := []string{"/assets/style.css", "/assets/app.js"}
	got := withoutAsset(plan, "assets/not-present.css")
	if len(got) != len(plan) {
		t.Errorf("got %v, want the plan unchanged", got)
	}
}
