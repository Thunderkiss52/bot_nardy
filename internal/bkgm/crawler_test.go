package bkgm

import (
	"strings"
	"testing"
)

func TestParseDocumentExtractsTextAndLinks(t *testing.T) {
	htmlDoc := `
<!doctype html>
<html>
<head><title>Backgammon Matches</title></head>
<body>
  <h1>Backgammon Matches</h1>
  <p>Annotated match with rollout comments.</p>
  <a href="/matches/woba.html">Woolsey vs Bagai</a>
  <a href="https://www.bkgm.com/articles/index.html">Articles</a>
  <script>ignore me</script>
</body>
</html>`

	doc, links, err := parseDocument("https://bkgm.com/matches/", []byte(htmlDoc))
	if err != nil {
		t.Fatalf("parseDocument failed: %v", err)
	}
	if doc.Title != "Backgammon Matches" {
		t.Fatalf("unexpected title: %q", doc.Title)
	}
	if doc.Category != "match" {
		t.Fatalf("unexpected category: %q", doc.Category)
	}
	if !strings.Contains(doc.Text, "Annotated match with rollout comments.") {
		t.Fatalf("expected body text, got %q", doc.Text)
	}
	if strings.Contains(doc.Text, "ignore me") {
		t.Fatalf("script text leaked into body: %q", doc.Text)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
}

func TestNormalizeURLFiltersAssetsAndExternalHosts(t *testing.T) {
	cases := []struct {
		raw string
		ok  bool
	}{
		{"https://bkgm.com/articles/index.html", true},
		{"https://www.bkgm.com/gloss/", true},
		{"https://bkgm.com/image.png", false},
		{"https://example.com/articles/", false},
	}

	for _, tc := range cases {
		_, ok := normalizeURL(tc.raw)
		if ok != tc.ok {
			t.Fatalf("normalizeURL(%q) ok=%v want %v", tc.raw, ok, tc.ok)
		}
	}
}
